package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var RootCmd = &cobra.Command{
	Use: "federal-sharing",
	Short: "FedSharing main net pool manager",
	Long: "FedSharing main net pool manager",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("FedSharing: Main-Blockchain v1.0")
	},
}

func Execute()  {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
