package cmd

import (
	"fedSharing/sidechain/configs"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var RootCmd = &cobra.Command{
	Use: configs.RootCmd,
	Short: configs.RootShort,
	Long: configs.RootLong,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("FedSharing: Side-Blockchain v1.0")
	},
}

func Execute()  {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}