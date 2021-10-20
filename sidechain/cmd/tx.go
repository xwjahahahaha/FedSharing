package cmd

import (
	"fedSharing/sidechain/blockchain"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/task"
	"fedSharing/sidechain/tx"
	"fedSharing/sidechain/utils"
	"fedSharing/sidechain/utxo"
	"fmt"
	"github.com/spf13/cobra"
	"math/big"
	"os"
	"strconv"
)

// Send
// 新增交易
var Send = &cobra.Command{
	Use: "send [from] [to] [amount]",
	Short: "make a new transaction",
	Long: "make a new transaction",
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		from, to := args[0], args[1]
		amount, _ := strconv.Atoi(args[2])
		bc := blockchain.LoadSideBlockChain("./database/", utils.AssemblySocketAddr(IP, NodePort))
		defer bc.LevelDB.Close()
		utxo := utxo.NewUTXO(bc)
		utxo.Reindex()
		ntx, err := utxo.NewTransaction(from, to, amount, "./wallet.dat")
		if err != nil {
			return err
		}
		exampleEpoch := task.GlobalEpoch{
			Idx:          0,
			Members: 	  map[string]float64{"a": 1, "b": 1, "c": 1},
			Contribution: make(map[string]*big.Float),
		}
		exampleTask := task.Task{
			Idx:          0,
			Epochs:       []*task.GlobalEpoch{&exampleEpoch},
			OriginatorID: 1,
		}
		coinbaseTx := tx.NewCoinbaseTx(from, "", int(configs.CoinBaseReward))
		nBlock, err := bc.AddNewBlock("", exampleTask, 0, []*tx.Transaction{coinbaseTx, ntx})
		if err != nil {
			return err
		}
		utxo.Update(nBlock)
		fmt.Fprintln(os.Stdout,"Success create a new transaction.")
		return nil
	},
}
