package blockchain

import (
	"bytes"
	"fedSharing/sidechain/block"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/logger"
	"fedSharing/sidechain/tx"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
)

type BlockChainIterator struct {
	currentHash []byte
	db *leveldb.DB
}

func (bci *BlockChainIterator) Next() *block.SideBlock {
	// 到达创世块停止
	if bytes.Compare(bci.currentHash, tx.InitialHash) == 0 {
		return nil
	}
	blockBytes, err := bci.db.Get(append([]byte(configs.BlockPrefix), bci.currentHash...), nil)
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	nowBlock := block.DeserializationBlock(blockBytes)
	// 更新currentHash
	bci.currentHash = nowBlock.Previous
	return nowBlock
}