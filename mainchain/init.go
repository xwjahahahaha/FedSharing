package main

import (
	cmd "fedSharing/mainchain/cmd"
	"fedSharing/mainchain/configs"
	"fmt"
	"github.com/ipfs/go-log/v2"
	"os"
)



func init() {
	initCmd()
	if err := configs.GlobalConfig.ReadConfigs(configs.FLConfFilePath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err := log.SetLogLevel("main-Blockchain", "Info")
	if err != nil {
		fmt.Println("SetLogLevel error.")
		return
	}
}

func initCmd()  {
	cmd.RootCmd.AddCommand(
		cmd.StartPoolManager,
		cmd.StartMiner,
	)
	cmd.RootCmd.PersistentFlags().StringVarP(&configs.FLConfFilePath, "fl-config-file", "c", "./configs/fl_conf.json", "联邦学习配置文件")
	cmd.RootCmd.MarkPersistentFlagRequired("fl-config-file")
	cmd.StartMiner.Flags().IntVarP(&configs.ClientID, "client-id", "i", 0, "客户端编号")
	cmd.StartMiner.MarkFlagRequired("client-id")
}
