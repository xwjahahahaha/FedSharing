package node

import (
	"context"
	"errors"
	"fedSharing/mainchain/configs"
	log2 "fedSharing/mainchain/log"
	"fedSharing/mainchain/measure"
	"fedSharing/mainchain/utils"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type MainChainNode struct {
	NodeInfo *Info
	Ctx context.Context
	Cancel context.CancelFunc
	Host host.Host
	NetWork *BlockchainNetwork
}

type BlockchainNetwork struct {

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	networkID string
	self peer.ID
}

type Info struct {
	Role Identity
	Id int
	Ip string
	Port int
}

type Identity int

const (
	PoolManager Identity = iota // value --> 0
	Miner
)

const StreamProtocol = "/fedSharing/1.0.0"

var (
	connectPeerChan = make(chan bool, 1)
	joinNetworkChan = make(chan bool, 1)
	DiffDeliveryOverChan = make(chan bool, 1)
	ModelDeliveryOverChan = make(chan bool, 1)
	WaitEpochChan chan []bool
	NewEpochMsgAlreadySend bool
	GlobalEpoch int
	ReceiveGlobalEpoch int
)


func NewHostNode(identity Identity, id int) (*MainChainNode, error) {
	if identity < 0 || identity > 1 {
		return nil, errors.New(" Not have this identity. ")
	}
	// 1.设置上下文环境
	ctx, cancel := context.WithCancel(context.Background())
	// 2.设置私钥
	privateKey, _, err := crypto.GenerateKeyPair(
		crypto.ECDSA,
		-1,
	)
	var newInfo Info
	switch identity {
	case PoolManager:
		newInfo = Info{
			Role: PoolManager,
			Id:   id,
			Ip:   configs.GlobalConfig.HostConfigViper.GetString("pool_manager.ip"),
			Port: configs.GlobalConfig.HostConfigViper.GetInt("pool_manager.port"),
		}
	case Miner:
		newInfo = Info{
			Role: Miner,
			Id:   id,
			Ip: configs.GlobalConfig.HostConfigViper.GetString("miner.ip"),
			Port: configs.GlobalConfig.HostConfigViper.GetInt("miner.port"),
		}
	}
	// 3.创建本地节点
	newHost, err := libp2p.New(ctx,
		// 使用自己生成的私钥
		libp2p.Identity(privateKey),
		//libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", newInfo.Ip, newInfo.Port)),
		//TODO
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/0")),
	)
	if err != nil {
		log2.Logger.Error(err)
		cancel()
		return nil, err
	}
	switch identity {
	case PoolManager:
		log2.Logger.Infof("\x1b[32m%s\x1b[0m", "Pool Manager Server")
	case Miner:
		log2.Logger.Infof("\x1b[32m%s%d\x1b[0m", "Miner Client ", id)
	}
	log2.Logger.Infof("\x1b[32m%s\x1b[0m", "Create local MainChain node success!")
	log2.Logger.Infof("\x1b[32m%s%s\x1b[0m", "Node Info : ", newHost.ID())
	log2.Logger.Infof("\x1b[32m%s\x1b[0m", newHost.Addrs())
	return &MainChainNode{
		NodeInfo: &newInfo,
		Ctx:      ctx,
		Cancel:   cancel,
		Host:     newHost,
		NetWork:  nil,
	}, nil
}

func (mcn *MainChainNode) StartNodeServer() error {
	// 1. 创建pubsub服务
	ps, err := pubsub.NewGossipSub(mcn.Ctx, mcn.Host)
	if err != nil {
		return err
	}
	// 2. 设置本地mDNS节点发现
	if err = setupDiscovery(mcn.Host, configs.GlobalConfig.HostConfigViper.GetString("network.mdns.discovery_service_tag")); err != nil {
		return err
	}
	<- connectPeerChan
	//3. 加入区块链网络
	netWork, err := JoinBlockchainNetWork(
		mcn,
		ps,
		configs.GlobalConfig.HostConfigViper.GetString("network.pubsub.id"),
	)
	if err != nil {
		return err
	}
	mcn.NetWork = netWork
	<- joinNetworkChan
	switch mcn.NodeInfo.Role {
	case PoolManager:
		// 发送模型
		ticker := time.NewTicker(5 * time.Second)
		for {
			<-ticker.C
			if isContinue := mcn.SendCollectClientsMsg(); isContinue {
				// 给最后连上的节点再发送一次
				mcn.SendCollectClientsMsg()
				break
			}
		}
		log2.Logger.Infof("\x1b[32m%s\x1b[0m", "All the Client already in place.Begin Federal Learning.")
		WaitEpochChan = make(chan []bool, 1)
		WaitEpochChan <- make([]bool, len(mcn.NetWork.ps.ListPeers(mcn.NetWork.topic.String())))
		GlobalEpoch += 1
		mcn.SendGlobalEpochMsg(GlobalEpoch)
	case Miner:
		fmt.Println("Miner 已经启动...")
	}
	WaitingCloseNode(mcn, mcn.NodeInfo.Role)
	return nil
}

func WaitingCloseNode(mcn *MainChainNode, identity Identity)  {
	// 等待SIGINT或SIGTERM信号
	ch := make(chan os.Signal, 1)
	// 当收到ctrl + c时将信号写入通道
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<- ch		// 等待阻塞
	// 关闭节点
	if err := mcn.Host.Close(); err != nil {
		log.Panic(err)
	}
	// 关闭上下文环境
	mcn.Cancel()
	// 将测量的数据写入文件
	switch identity {
	case PoolManager:
		measure.WriteMeasureTimeToFile("./measure/out/time_server.json")
	case Miner:
		measure.WriteMeasureTimeToFile("./measure/out/time_client_" + strconv.Itoa(configs.ClientID) + ".json")
	}
	utils.ColorPrint("Received signal, shutting down...")
	utils.ColorPrint("ByeBye~")
}


