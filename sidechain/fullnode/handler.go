package fullnode

import (
	"bytes"
	"encoding/gob"
	"fedSharing/sidechain/block"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/logger"
	"fedSharing/sidechain/utils"
	"fedSharing/sidechain/utxo"
	"strconv"
)

// handleNewMsg
// @Description: 新消息路由
// @receiver bn
// @param msgData
// @param bc
func (bn *BlockchainNetwork) handleNewMsg(msgData []byte, utxo *utxo.UnSpendTxOutputSet)  {
	if len(msgData) <= utils.CommandLength {
		return
	}
	command := utils.BytesToCommand(msgData[:utils.CommandLength])
	logger.Logger.Infof("\x1b[32mReceived %s command\x1b[0m", command)
	switch command {
	case "version":
		bn.handleVersionMsg(msgData[utils.CommandLength:], utxo)
	case "getblock":
		bn.handleGetBlockMsg(msgData[utils.CommandLength:], utxo)
	case "block":
		bn.handleBlockMsg(msgData[utils.CommandLength:], utxo)
	default:
		logger.Logger.Warn("Command is not matched.")
	}
}

// handleVersionMsg
// @Description: 版本消息处理函数
// @receiver bn
// @param payload
// @param bc
func (bn *BlockchainNetwork) handleVersionMsg(payload []byte, utxo *utxo.UnSpendTxOutputSet)  {
	var vm VersionMsg
	var buff bytes.Buffer
	buff.Write(payload)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&vm)
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	if vm.SideChainVersion > configs.ChainVersion {
		logger.Logger.Warn("your side-blockchain version is old, please update first.")
		return
	}
	// 最长链规则
	localBestHeight := utxo.SBC.GetBestHeight()
	if localBestHeight < vm.BestHeight {
		// 本地版本小，则更新版本
		bn.SendGetBlocks(localBestHeight+1, vm.BestHeight)
	}else if localBestHeight > vm.BestHeight{
		//fmt.Println("我需要给他发送新区块") //TODO 我需要给他发送新区块?
	}
	logger.Logger.Infof("[%s]'s version:%s,height:%s", vm.AddrFrom, strconv.Itoa(vm.SideChainVersion), strconv.Itoa(vm.BestHeight))
}

// handleGetBlockMsg
// @Description: 获取区块消息处理函数
// @receiver bn
// @param payload
// @param bc
func (bn *BlockchainNetwork) handleGetBlockMsg(payload []byte, utxo *utxo.UnSpendTxOutputSet)  {
	var gb GetBlockMsg
	var buff bytes.Buffer
	buff.Write(payload)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&gb)
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	// 获取本地所有区块Hash
	hashes := utxo.SBC.GetBlockHashes()
	for height := gb.HeightRange[0]; height <= gb.HeightRange[1]; height++ {
		if blockHash, has := hashes[int64(height)]; has {
			bn.SendBlock(utxo.SBC.GetBlock(blockHash))
		}else {
			logger.Logger.Warnf("There is no this block (heigth=%d).", height)
		}
	}
}

func (bn *BlockchainNetwork) handleBlockMsg(payload []byte, utxo *utxo.UnSpendTxOutputSet)  {
	var bm BlockMsg
	var buff bytes.Buffer
	buff.Write(payload)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&bm)
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	newBlock := block.DeserializationBlock(bm.Block)
	// 验证区块, 成功则加入区块链
	if _, err := utxo.SBC.LinkNewBlock(newBlock); err != nil {
		logger.Logger.Warn(err)
		return
	}
	// 更新UTXO
	utxo.Update(newBlock)
}

