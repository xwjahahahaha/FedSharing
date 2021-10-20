package main

import (
	"fedSharing/sidechain/cmd"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/fullnode"
	"gopkg.in/ini.v1"
	"log"
)

func init()  {

	// 初始化配置文件
	initConfigs("./configs/config.ini")

	// 初始化命令行
	initCmd()

}

// initConfigs
// @Description: 加载配置文件
// @param path	配置文件路径
func initConfigs(path string)  {
	file, err := ini.Load(path)
	if err != nil {
		log.Panic("配置文件读取错误，请检查配置文件！")
	}
	configs.LoadBlockChainConfig(file)
	configs.LoadConsensus(file)
	configs.LoadDataBase(file)
	configs.LoadCmdConfig(file)
	configs.LoadTransaction(file)
	configs.LoadEnCrypto(file)
	configs.LoadBlock(file)
	configs.LoadBlockchain(file)
}

func initCmd()  {
	cmd.RootCmd.AddCommand(
		cmd.CreateBlockChain,
		cmd.BlockChainPrint,
		cmd.Send,
		cmd.GetBalance,
		cmd.CreateWallet,
		cmd.StartNode,
		cmd.LoadWallet,
	)
	// 添加一些flag
	AddSomeBCFlags()
	AddAndResolveNodeFlag()
}

// AddSomeBCFlags
// @Description: 区块链cmd中添加一些flag
func AddSomeBCFlags()  {
	cmd.CreateBlockChain.Flags().IntVarP(&cmd.NodePort, "node port", "p", 1902, "节点端口号")
	cmd.CreateBlockChain.Flags().StringVarP(&cmd.IP, "node IP address", "i", "127.0.0.1", "节点IP地址")
	cmd.CreateBlockChain.Flags().StringVarP(&cmd.Address, "initAddress", "a", cmd.DefaultAddress, "初始化账户地址")
	cmd.BlockChainPrint.Flags().IntVarP(&cmd.NodePort, "node port", "p", 1902, "节点端口号")
	cmd.BlockChainPrint.Flags().StringVarP(&cmd.IP, "node IP address", "i", "127.0.0.1", "节点IP地址")
	cmd.GetBalance.Flags().IntVarP(&cmd.NodePort, "node port", "p", 1902, "节点端口号")
	cmd.GetBalance.Flags().StringVarP(&cmd.IP, "node IP address", "i", "127.0.0.1", "节点IP地址")
	cmd.CreateWallet.Flags().StringVarP(&cmd.WalletFilePath, "wallet .dat file path", "f", "./wallet.dat", "钱包文件路径")
	cmd.Send.Flags().StringVarP(&cmd.IP, "node IP address", "i", "127.0.0.1", "节点IP地址")
	cmd.Send.Flags().IntVarP(&cmd.NodePort, "node port", "p", 1902, "节点端口号")
}


func AddAndResolveNodeFlag()  {
	cmd.NodeConfigs = fullnode.NodeConfig{}
	cmd.StartNode.Flags().StringVar(&cmd.NodeConfigs.NetWorkID, "network", "FedSharing_SideNetWork",
		"SideChainP2P网络ID")
	cmd.StartNode.Flags().StringVar(&cmd.NodeConfigs.DiscoveryServiceTag, "DiscoveryServiceTag", "FedSharing_Peers", "侧链节点发现标识符")
	cmd.StartNode.Flags().StringVarP(&cmd.NodeConfigs.IP, "node IP address", "i", "127.0.0.1", "节点IP地址")
	cmd.StartNode.Flags().IntVarP(&cmd.NodeConfigs.Port, "node port", "p", 1902, "节点端口号")
	//cmd.StartNode.Flags().Var(&cmd.NodeConfigs.BootstrapPeers, "peer", "向当前节点添加一组引导节点数组（mutiaddress格式）")
	//cmd.StartNode.Flags().VarP(&cmd.NodeConfigs.ListenAddresses, "listen", "l","向当前节点添加一组监听节点数组（mutiaddress格式）")
	//cmd.StartNode.Flags().StringVar(&cmd.NodeConfigs.ProtocolID, "pid", "/SideChain_P2P/1.0.0", "侧链协议号")
}