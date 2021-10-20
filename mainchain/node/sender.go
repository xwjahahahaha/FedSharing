package node

import (
	"fedSharing/mainchain/configs"
	"fedSharing/mainchain/log"
	"fedSharing/mainchain/utils"
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
)

var LastConfirm bool

func (mcn *MainChainNode) Publish(data []byte) {
	if len(data) == 0 {
		log.Logger.Error("Can't push empty message data.")
		return
	}
	if err := mcn.NetWork.topic.Publish(mcn.NetWork.ctx, data); err != nil {
		log.Logger.Error(err)
		return
	}
	log.Logger.Infof("Success push %s message.", data[:utils.CommandLength])
}

func (mcn *MainChainNode) SendCollectClientsMsg() bool {
	peerList := mcn.NetWork.ps.ListPeers(mcn.NetWork.topic.String())
	count := len(peerList)
	if count == configs.GlobalConfig.FlConfigViper.GetInt("clients_num") {
		if LastConfirm {
			return true
		}
		LastConfirm = true
	}
	peerStringList := make([]string, 0)
	for i := 0; i < count; i++ {
		peerStringList = append(peerStringList, peerList[i].Pretty())
	}
	newCCM := &CollectClientsMsg{
		TaskID:         configs.GlobalConfig.FlConfigViper.GetInt("task.id"),
		NowClientCount: count,
		NowClients:     peerStringList,
		AddrFrom:       mcn.NetWork.self.Pretty(),
	}
	log.Logger.Infof("\x1b[32m\nTaskId:%d\nPoolManagerServer:%s\nClientsCount:%d\n%v\x1b[0m",
		newCCM.TaskID,
		newCCM.AddrFrom,
		newCCM.NowClientCount,
		newCCM.NowClients,
	)
	mcn.Publish(append(utils.CommandToBytes("collect-clients"), utils.GobEncode(newCCM)...))
	return false
}

func (mcn *MainChainNode) SendGlobalEpochMsg(globalEpoch int) {
	utils.ColorPrint("=========================================================================")
	utils.ColorPrint(fmt.Sprintf("New global epoch start: [%d]", globalEpoch))
	mcn.Publish(append(utils.CommandToBytes("global-epoch"),
		utils.GobEncode(&GlobalEpochMsg{
			GlobalEpoch: globalEpoch,
			AddrFrom:    mcn.NetWork.self.Pretty(),
		})...))
	if globalEpoch == 1 {
		// TODO 部署合约
		utils.ColorPrint("Ready to push smart contract.")
	}
}

func (mcn *MainChainNode) SendEstablishStreamMsg(handler network.StreamHandler, command string) {
	newESM := EstablishStreamMsg{
		MultiAddr: mcn.Host.Addrs()[0].String(),
		ClientID: configs.ClientID,
		AddrFrom:  mcn.NetWork.self.Pretty(),
	}
	mcn.Publish(append(utils.CommandToBytes(command), utils.GobEncode(newESM)...))
	// 设置流处理函数
	mcn.Host.SetStreamHandler(StreamProtocol, handler)
}

