package fullnode

import (
	"fedSharing/sidechain/block"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/logger"
	"fedSharing/sidechain/utils"
)

// Publish
// @Description: 向topic发布一个消息数据
// @receiver bn
// @param data
func (bn *BlockchainNetwork) Publish(data []byte) {
	if len(data) == 0 {
		logger.Logger.Error("Can't push empty message data.")
		return
	}
	if err := bn.topic.Publish(bn.ctx, data); err != nil {
		logger.Logger.Error(err)
		return
	}
	logger.Logger.Infof("Success push %s message.", data[:utils.CommandLength])
}

// SendVersionMsg
// @Description: 发布当前消息数据
// @receiver bn
// @param height
func (bn *BlockchainNetwork) SendVersionMsg(height int)  {
	newVersion := &VersionMsg{
		SideChainVersion: configs.ChainVersion,
		BestHeight:       height,
		AddrFrom:         bn.self.Pretty(),
	}
	bn.Publish(append(utils.CommandToBytes("version"), utils.GobEncode(newVersion)...))
}

// SendGetBlocks
// @Description: 发布区块请求消息
// @receiver bn
// @param start
// @param end
// @param addr
func (bn *BlockchainNetwork) SendGetBlocks(start, end int)  {
	if start > end {
		logger.Logger.Warn("Block request range error.")
		return
	}
	newGetBlockVersion := &GetBlockMsg{
		HeightRange: []int{start, end},
		AddrFrom:   bn.self.Pretty(),
	}
	bn.Publish(append(utils.CommandToBytes("getblock"), utils.GobEncode(newGetBlockVersion)...))
}

// SendBlock
// @Description: 发送区块消息
// @receiver bn
// @param block
func (bn *BlockchainNetwork) SendBlock(block *block.SideBlock)  {
	newBlockMsg := BlockMsg{
		Block:    block.Serialization(),
		AddrFrom: bn.self.Pretty(),
	}
	bn.Publish(append(utils.CommandToBytes("block"), utils.GobEncode(newBlockMsg)...))
}