package node

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"fedSharing/mainchain/configs"
	"fedSharing/mainchain/execCmd"
	"fedSharing/mainchain/log"
	"fedSharing/mainchain/measure"
	"fedSharing/mainchain/utils"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	mutiaddr "github.com/multiformats/go-multiaddr"
	"os"
	"strconv"
	"time"
)

var PoolManagerPeerID string

func (mcn *MainChainNode) handleNewMsg(msg *pubsub.Message, identity Identity) {
	if len(msg.Data) <= utils.CommandLength {
		log.Logger.Warnf("Receive invalid msg data.")
		return
	}
	command := utils.BytesToCommand(msg.Data[:utils.CommandLength])
	utils.ColorPrint(fmt.Sprintf("Received 【%s】 command", command))
	switch command {
	case "collect-clients":
		if identity == Miner {
			mcn.handleCollectClients(msg.Data[utils.CommandLength:])
		}
	case "global-epoch":
		if identity == Miner {
			mcn.handleGlobalEpoch(msg.Data[utils.CommandLength:])
		}
	case "establish-diff-stream":
		if identity == PoolManager {
			mcn.handleEstablishStream(msg.Data[utils.CommandLength:], 1)
		}
	case "establish-model-stream":
		if identity == PoolManager {
			mcn.handleEstablishStream(msg.Data[utils.CommandLength:], 2)
		}
	default:
		log.Logger.Warn("Command is not matched.")
		return
	}
}

func (mcn *MainChainNode) handleCollectClients(payload []byte) {
	var ccm CollectClientsMsg
	var buf bytes.Buffer
	buf.Write(payload)
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(&ccm); err != nil {
		log.Logger.Error(err)
		return
	}
	utils.ColorPrint(fmt.Sprintf("\nTaskId:%d\nPoolManagerServer:%s\nClientsCount:%d\n%v",
		ccm.TaskID,
		ccm.AddrFrom,
		ccm.NowClientCount,
		ccm.NowClients))
	// 设置当前Pool Manager
	if len(PoolManagerPeerID) == 0 {
		PoolManagerPeerID = ccm.AddrFrom
	}
}

func (mcn *MainChainNode) handleGlobalEpoch(payload []byte) {
	var gem GlobalEpochMsg
	var buf bytes.Buffer
	buf.Write(payload)
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(&gem); err != nil {
		log.Logger.Error(err)
		return
	}
	utils.ColorPrint(fmt.Sprintf("Start global epoch [%d]", gem.GlobalEpoch))
	ReceiveGlobalEpoch = gem.GlobalEpoch
	if gem.GlobalEpoch > 1 {
		// 发送更新模型数据流申请, 并读取模型
		mcn.SendEstablishStreamMsg(handleModelStream, "establish-model-stream")
		<- ModelDeliveryOverChan
		// 更新本地模型
	}
	start := time.Now()
	// 本地训练并保存diff与最新本地模型
	modelSavePath := configs.GlobalConfig.FlConfigViper.GetString("model_path") + "MinerClient_" + strconv.Itoa(configs.ClientID) + "/"
	diffSavePath := configs.GlobalConfig.FlConfigViper.GetString("diff_path") + "MinerClient_" + strconv.Itoa(configs.ClientID) + "/"
	err := execCmd.CmdAndChangeDirToShow("./", "python", []string{"./python_fl/client.py", "-f", "2",
		"-c", configs.FLConfFilePath,
		"-m", modelSavePath,
		"-d", diffSavePath,
		"-e", strconv.Itoa(gem.GlobalEpoch),
		"-i", strconv.Itoa(configs.ClientID),
	})
	if err != nil {
		log.Logger.Error(err)
		return
	}
	measure.MeasureTime("local_train" + ".epoch_" + strconv.Itoa(gem.GlobalEpoch) + ".client_" + strconv.Itoa(configs.ClientID), start)
	// 与server建立stream链接
	mcn.SendEstablishStreamMsg(handleDiffStream, "establish-diff-stream")
}

func (mcn *MainChainNode) handleEstablishStream(payload []byte, kind int) {
	var esm EstablishStreamMsg
	var buf bytes.Buffer
	buf.Write(payload)
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(&esm); err != nil {
		log.Logger.Error(err)
		return
	}
	// 提取目标节点的Peer ID信息
	desMultiaddr, err := mutiaddr.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", esm.MultiAddr, esm.AddrFrom))
	if err != nil {
		log.Logger.Error(err)
		return
	}
	info, err := peer.AddrInfoFromP2pAddr(desMultiaddr)
	if err != nil {
		log.Logger.Error(err)
		return
	}
	mcn.Host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	// 建立流
	newStream, err := mcn.Host.NewStream(context.Background(), info.ID, StreamProtocol)
	if err != nil {
		log.Logger.Error(err)
		return
	}
	utils.ColorPrint("Success establish new stream: " + newStream.ID() + " from client " + strconv.Itoa(esm.ClientID))
	rw := bufio.NewReadWriter(bufio.NewReader(newStream), bufio.NewWriter(newStream))
	if kind == 1 {
		// 建立更新diff的流
		mcn.handleDiffStream(rw, esm.ClientID)
	} else {
		// 建立更新模型的流
		go writeModelData(rw, esm.ClientID)
	}
}

func (mcn *MainChainNode) handleDiffStream(rw *bufio.ReadWriter, clientId int) {
	dirRootPath := configs.GlobalConfig.FlConfigViper.GetString("diff_path") +
		"ServerReceived/Client_" + strconv.Itoa(clientId) + "/"
	diffFilePath := dirRootPath + "diff_epoch_" + strconv.Itoa(GlobalEpoch) + ".dict"
	if !utils.DirectoryOrFileExists(dirRootPath) {
		err := os.MkdirAll(dirRootPath, os.ModePerm)
		if err != nil {
			log.Logger.Error(err)
			return
		}
	}
	go readDiffData(rw, diffFilePath)
	<- DiffDeliveryOverChan
	// 读取之后python聚合
	start := time.Now()
	modelSavePath := configs.GlobalConfig.FlConfigViper.GetString("model_path") + "PoolManagerServer/"
	err := execCmd.CmdAndChangeDirToShow("./", "python", []string{"./python_fl/server.py", "-f", "2",
		"-c", configs.FLConfFilePath,
		"-m", modelSavePath,
		"-d", diffFilePath,
	})
	if err != nil {
		log.Logger.Error(err)
		return
	}
	measure.MeasureTime("server_aggregate" + ".epoch_" + strconv.Itoa(GlobalEpoch) + ".client_" + strconv.Itoa(clientId) , start)
	// 评估效果
	start = time.Now()
	err = execCmd.CmdAndChangeDirToShow("./", "python", []string{"./python_fl/server.py", "-f", "3",
		"-c", configs.FLConfFilePath,
		"-m", modelSavePath,
	})
	if err != nil {
		log.Logger.Error(err)
		return
	}
	measure.MeasureTime("server_assess" + ".epoch_" + strconv.Itoa(GlobalEpoch) + ".client_" + strconv.Itoa(clientId) , start)
	clientsAry := <- WaitEpochChan
	clientsAry[clientId] = true
	for clientId, over := range clientsAry {
		if !over {
			utils.ColorPrint(fmt.Sprintf("Waitting for Client %d", clientId))
			WaitEpochChan <- clientsAry
			return
		}
	}
	NewEpochMsgAlreadySend = false
	WaitEpochChan <- make([]bool, len(mcn.NetWork.ps.ListPeers(mcn.NetWork.topic.String())))
	if GlobalEpoch < configs.GlobalConfig.FlConfigViper.GetInt("global_epochs"){
		if !NewEpochMsgAlreadySend {
			// 开启新的epoch
			GlobalEpoch += 1
			mcn.SendGlobalEpochMsg(GlobalEpoch)
			NewEpochMsgAlreadySend = true
		}
		return
	}
	utils.ColorPrint(fmt.Sprintf("FedSharing Task [%d] is completed.", configs.GlobalConfig.FlConfigViper.GetInt("task.id")))
}


