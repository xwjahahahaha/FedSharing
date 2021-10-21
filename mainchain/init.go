package main

import (
	cmd "fedSharing/mainchain/cmd"
	"fedSharing/mainchain/configs"
	"fedSharing/mainchain/measure"
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
	initMeasure()
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

func initMeasure()  {
	globalEpoch := configs.GlobalConfig.FlConfigViper.GetInt("global_epochs")
	measure.MeasureStruct = make(map[string][]map[string]int64)
	measure.MeasureStruct[measure.LOCALTRAIN] = make([]map[string]int64, globalEpoch+1)
	measure.MeasureStruct[measure.AGGREGATE] = make([]map[string]int64, globalEpoch+1)
	measure.MeasureStruct[measure.DIFF] = make([]map[string]int64, globalEpoch+1)
	measure.MeasureStruct[measure.ASSESS] = make([]map[string]int64, globalEpoch+1)
	measure.MeasureStruct[measure.SENDMODEL] = make([]map[string]int64, globalEpoch+1)
	for k := range measure.MeasureStruct {
		for i:= range measure.MeasureStruct[k] {
			measure.MeasureStruct[k][i] = make(map[string]int64)
		}
	}
}

