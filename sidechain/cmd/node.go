package cmd

import (
	"fedSharing/sidechain/fullnode"
	"github.com/spf13/cobra"
)



var (
	IP 			string		// IP地址
	NodePort 	int			// 端口号
	NodeConfigs  fullnode.NodeConfig	// 节点配置
)



// StartNode
// 开启本地节点
var StartNode = &cobra.Command{
	Use: "start",
	Short: "start your node server",
	Long: "start your node server",
	Args: cobra.ExactArgs(0), // TODO
	Run: func(cmd *cobra.Command, args []string) {
		// 1.加载区块链
		//bc := blockchain.LoadSideBlockChain("./database/", utils.AssemblySocketAddr(NodeConfigs.IP, NodeConfigs.Port))
		//defer bc.LevelDB.Close()
		// 2.创建节点、启动节点
		scn := fullnode.NewHostNode(NodeConfigs)
		defer scn.Cancel()
		scn.StartNodeServer()

	},
}

