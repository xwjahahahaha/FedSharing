package cmd

import (
	"fedSharing/sidechain/blockchain"
	"fedSharing/sidechain/utils"
	"fedSharing/sidechain/utxo"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// CreateBlockChain
// 创建区块链
var	CreateBlockChain = &cobra.Command{
	Use: "crtBC",
	Short: "create your side-blockchain",
	Long: "create your side-blockchain",
	RunE: func(cmd *cobra.Command, args []string) error {
		if Address == DefaultAddress {
			// 创建一个空区块链
			nebc, err := blockchain.NewEmptySideBlockChain("./database/", utils.AssemblySocketAddr(IP, NodePort))
			if err != nil {
				return err
			}
			defer nebc.LevelDB.Close()
			fmt.Fprintln(os.Stdout, "New empty blockchain created!")
		}else {
			// 创建一个初始化区块链
			// 1. 创建区块链
			nbc, err := blockchain.NewSideBlockChain(Address, "./database/", utils.AssemblySocketAddr(IP, NodePort))
			if err != nil {
				return err
			}
			defer nbc.LevelDB.Close()
			// 2. 更新存储UTXO集合
			utxo := utxo.NewUTXO(nbc)
			utxo.Reindex()
			fmt.Fprintln(os.Stdout, "New blockchain created!")
		}
		return nil
	},
}


// BlockChainPrint
// 输出区块链
var	BlockChainPrint = &cobra.Command{
	Use: "prtBC",
	Short: "print your blockchain data",
	Long: "print your blockchain data",
	Run: func(cmd *cobra.Command, args []string) {
		bc := blockchain.LoadSideBlockChain("./database/", utils.AssemblySocketAddr(IP, NodePort))
		defer bc.LevelDB.Close()
		utxo := utxo.NewUTXO(bc)
		utxo.Reindex()
		bc.StdOutputSideBlockChain()
	},
}





