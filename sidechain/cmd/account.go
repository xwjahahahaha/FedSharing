package cmd

import (
	"errors"
	"fedSharing/sidechain/account"
	"fedSharing/sidechain/blockchain"
	"fedSharing/sidechain/utils"
	"fedSharing/sidechain/utxo"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const DefaultAddress = "FS_000000000000000000000000000000000"

var (
	Address string			// 账户地址
	WalletFilePath string	// 钱包文件路径
)

// CreateWallet
// 创建一个钱包
var CreateWallet = &cobra.Command{
	Use: "crtWal",
	Short: "Create a wallet with 10 random addresses and store it locally permanently.",
	Long: "Create a wallet with 10 random addresses and store it locally permanently.",
	Run: func(cmd *cobra.Command, args []string) {
		w := account.InitRandomWallet()
		fmt.Fprintln(os.Stdout, w.String())
		w.SaveToFile(WalletFilePath)		// 不选择就默认"./wallet.dat"
		fmt.Fprintln(os.Stdout,"Success create a new random wallet.")
	},
}

// LoadWallet
// 加载钱包
var LoadWallet = &cobra.Command{
	Use: "lodWal [walletFilePath]",
	Short: "Loads the wallet from a local file.",
	Long: "Loads the wallet from a local file.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		w := account.LoadLocalWalletFile(args[0])
		fmt.Fprintln(os.Stdout, w.String())
	},
}

// GetBalance
// 获取余额
var GetBalance = &cobra.Command{
	Use: "getBal [address]",
	Short: "get your balance",
	Long: "get your balance",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		address := args[0]
		if !account.ValidateAddress(address) {
			return errors.New(" Not a valid address.")
		}
		bc := blockchain.LoadSideBlockChain("./database/", utils.AssemblySocketAddr(IP, NodePort))
		defer bc.LevelDB.Close()
		utxo := utxo.NewUTXO(bc)
		utxo.Reindex()
		outputs := utxo.FindUTXOByOne(account.ResolveAddressToPubKeyHash(address))
		amount := 0
		for _, output := range outputs.Outputs {
			amount += output.Value
		}
		fmt.Fprintln(os.Stdout, "----[", address, "]的余额为：", amount)
		return nil
	},
}