package node

import (
	"fedSharing/mainchain/configs"
	"fedSharing/mainchain/log"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)


func JoinBlockchainNetWork(mcn *MainChainNode, ps *pubsub.PubSub, networkID string) (*BlockchainNetwork, error) {
	topic, err := ps.Join(configs.GlobalConfig.HostConfigViper.GetString("network.pubsub.topic"))
	if err != nil {
		return nil, err
	}
	log.Logger.Infof("success join main-blockchain network %s.", networkID)
	sub, err := topic.Subscribe(pubsub.WithBufferSize(1024*10))		// 订阅网络
	if err != nil {
		return nil, err
	}
	log.Logger.Infof("success subscribe in the %s network.", networkID)
	bn := &BlockchainNetwork{
		ctx:         mcn.Ctx,
		ps:          ps,
		topic:       topic,
		sub:         sub,
		networkID:   networkID,
		self:        mcn.Host.ID(),
	}
	mcn.NetWork = bn
	// 开启循环读取
	go mcn.readLoop(mcn.NodeInfo.Role)
	return bn, nil
}

func (mcn *MainChainNode) readLoop(identity Identity)  {
	joinNetworkChan <- true
	for {
		msg, err := mcn.NetWork.sub.Next(mcn.NetWork.ctx)
		if err != nil {
			return
		}
		// 避免自循环
		if msg.ReceivedFrom == mcn.NetWork.self{
			continue
		}
		// client屏蔽所有其他peer
		if identity == Miner && len(PoolManagerPeerID)>0 && msg.ReceivedFrom.Pretty() != PoolManagerPeerID {
			continue
		}
		// 处理新msg
		mcn.handleNewMsg(msg, identity)
	}

}

