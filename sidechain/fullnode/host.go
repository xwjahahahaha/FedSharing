package fullnode

import (
	"context"
	"fedSharing/sidechain/blockchain"
	"fedSharing/sidechain/logger"
	"fedSharing/sidechain/utils"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"os"
	"time"
)

var (
	connectPeerChan = make(chan peer.ID, 1)
	joinNetworkChan = make(chan bool, 1)
	bcLevelDBChan = make(chan bool, 1)
)

// NodeConfig 初始启动配置
type NodeConfig struct {
	IP string
	Port int
	NetWorkID string
	DiscoveryServiceTag string
}

type SideChainNode struct {
	Ctx context.Context
	Cancel context.CancelFunc
	Node host.Host
	NodeConfig *NodeConfig
	NetWork *BlockchainNetwork
}

// NewHostNode
// @Description: 新建一个本地区块链节点
// @param config
// @return *SideChainNode
func NewHostNode(config NodeConfig) *SideChainNode {
	// 1.设置上下文环境
	ctx, cancel := context.WithCancel(context.Background())
	// 2.设置私钥 TODO 与本地区块链的公私钥体系统一
	privateKey, _, err := crypto.GenerateKeyPair(
		crypto.ECDSA,
		-1,
	)
	// 3.创建本地节点
	node, err := libp2p.New(ctx,
		// 使用自己生成的私钥
		libp2p.Identity(privateKey),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),		// 使用随机端口
	)
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	logger.Logger.Infof("\x1b[32m%s\x1b[0m", "Create local SideChain node success!")
	logger.Logger.Infof("\x1b[32m%s%s\x1b[0m", "Node Info : ", node.ID())
	logger.Logger.Infof("\x1b[32m%s\x1b[0m", node.Addrs())
	return &SideChainNode{
		Ctx:    ctx,
		Cancel: cancel,
		Node:   node,
		NodeConfig: &config,
	}
}

// StartNodeServer
// @Description: 启动本地节点服务
// @receiver scn
func (scn *SideChainNode) StartNodeServer() {
	// 1. 创建pubsub服务
	ps, err := pubsub.NewGossipSub(scn.Ctx, scn.Node)
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	// 2. 设置本地mDNS节点发现
	if err = setupDiscovery(scn.Node, scn.NodeConfig.DiscoveryServiceTag); err != nil {
		logger.Logger.Error(err)
		return
	}
	<- connectPeerChan
	//3. 加入区块链网络
	netWork, err := JoinBlockchainNetWork(scn.Ctx, ps, scn.Node.ID(), scn.NodeConfig.NetWorkID, utils.AssemblySocketAddr(scn.NodeConfig.IP, scn.NodeConfig.Port))
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	scn.NetWork = netWork
	<- joinNetworkChan
	// 定时同步版本信息
	ticker := time.NewTicker(5 * time.Second)
	go tickerSynVersion(netWork, ticker, utils.AssemblySocketAddr(scn.NodeConfig.IP, scn.NodeConfig.Port))
	select{}
}

// tickerSynVersion
// @Description: 定时同步版本信息
// @param network
// @param bc
// @param ticker
func tickerSynVersion(network *BlockchainNetwork, ticker *time.Ticker, socketAddr string)  {
	for {
		<- ticker.C
		<- bcLevelDBChan
		logger.Logger.Info("Send version synchronization information.")
		bc := blockchain.LoadSideBlockChain("./database/", socketAddr)
		height := bc.GetBestHeight()
		bc.LevelDB.Close()
		bcLevelDBChan <- true
		network.SendVersionMsg(height)
	}
}

