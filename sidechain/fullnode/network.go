package fullnode

import (
	"context"
	"fedSharing/sidechain/blockchain"
	"fedSharing/sidechain/logger"
	"fedSharing/sidechain/utxo"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

//const VersionChanBuffSize = 100

type BlockchainNetwork struct {
	//receiveBlocksChan chan *block.SideBlock		// 接受到的区块

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	networkID string
	self peer.ID
}

// JoinBlockchainNetWork
// @Description: 加入侧链网络
// @param ctx
// @param ps
// @param selfID
// @param networkID
// @param socketAddr
// @return *BlockchainNetwork
// @return error
func JoinBlockchainNetWork(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, networkID string, socketAddr string) (*BlockchainNetwork, error) {
	topic, err := ps.Join(networkID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	logger.Logger.Infof("success join side-blockchain network %s.", networkID)
	sub, err := topic.Subscribe()		// 订阅网络
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	logger.Logger.Infof("success subscribe in the %s network.", networkID)
	bn := &BlockchainNetwork{
		//receiveBlocksChan: make(chan *block.SideBlock, 0),
		ctx:         ctx,
		ps:          ps,
		topic:       topic,
		sub:         sub,
		networkID:   networkID,
		self:        selfID,
	}
	// 开启循环读取
	go bn.readLoop(socketAddr)
	return bn, nil
}

// readLoop
// @Description: 循环读取网络订阅消息
// @receiver bn
func (bn *BlockchainNetwork) readLoop(socketAddr string)  {
	joinNetworkChan <- true
	bcLevelDBChan <- true
	for {
		msg, err := bn.sub.Next(bn.ctx)
		if err != nil {
			//close(bn.VersionChan)
			return
		}
		// 避免自循环
		if msg.ReceivedFrom == bn.self {
			continue
		}
		// 加载本地数据库
		<- bcLevelDBChan
		bc := blockchain.LoadSideBlockChain("./database/", socketAddr)
		newUTXO := utxo.NewUTXO(bc)
		newUTXO.Reindex()
		// 处理新msg
		bn.handleNewMsg(msg.Data, newUTXO)
		bc.LevelDB.Close()
		bcLevelDBChan <- true
	}
}

