package cmd

import (
	"bytes"
	"errors"
	"fedSharing/mainchain/configs"
	"fedSharing/mainchain/node"
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
	"strconv"
)

var StartPoolManager = &cobra.Command{
	Use: "start-pool-manager",
	Short: "start your pool manager server",
	Long: "start your pool manager server",
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 初始化server
		modelSavePath := configs.GlobalConfig.FlConfigViper.GetString("model_path") + "PoolManagerServer/"
		commandExec := exec.Command("python", "./python_fl/server.py",
			"-f", "1", "-c", configs.FLConfFilePath, "-m", modelSavePath)
		commandExec.Stdout = &bytes.Buffer{}
		commandExec.Stderr = &bytes.Buffer{}
		err := commandExec.Run()
		if err != nil{
			fmt.Println(err)
			fmt.Println(commandExec.Stderr.(*bytes.Buffer).String())
			return err
		}
		fmt.Println(commandExec.Stdout.(*bytes.Buffer).String())
		// 创建p2p节点
		hostNode, err := node.NewHostNode(node.PoolManager, 0)
		if err != nil {
			return err
		}
		// 启动节点
		if err := hostNode.StartNodeServer(); err != nil {
			return err
		}
		return nil
	},
}

var StartMiner = &cobra.Command{
	Use: "start-miner",
	Short: "start your main network miner(FL-Client)",
	Long: "start your main network miner(FL-Client)",
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if configs.ClientID < 0 || configs.ClientID >= configs.GlobalConfig.FlConfigViper.GetInt("clients_num") {
			return errors.New(" Invalid client id. ")
		}
		modelSavePath := configs.GlobalConfig.FlConfigViper.GetString("model_path") + "MinerClient_" + strconv.Itoa(configs.ClientID) + "/"
		commandExec := exec.Command("python", "./python_fl/client.py",
			"-f", "1", "-c", configs.FLConfFilePath, "-m", modelSavePath)
		commandExec.Stdout = &bytes.Buffer{}
		commandExec.Stderr = &bytes.Buffer{}
		err := commandExec.Run()
		if err != nil{
			fmt.Println(err)
			fmt.Println(commandExec.Stderr.(*bytes.Buffer).String())
			return err
		}
		fmt.Println(commandExec.Stdout.(*bytes.Buffer).String())
		hostNode, err := node.NewHostNode(node.Miner, configs.ClientID)
		if err != nil {
			return err
		}
		if err := hostNode.StartNodeServer(); err != nil {
			return err
		}
		return nil
	},
}
