package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	sideBlock "fedSharing/sidechain/block"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/logger"
	"fedSharing/sidechain/task"
	"fedSharing/sidechain/tx"
	"fedSharing/sidechain/utils"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
)

type Blockchain struct {
	LatestHash []byte
	LevelDB *leveldb.DB
}

// NewSideBlockChain
// @Description: 创建新的侧链
// @param addr
// @param nodeInfo
// @return *Blockchain
func NewSideBlockChain(addr string, dbPath string, nodeInfo string) (*Blockchain, error) {
	var newBC Blockchain
	// 1. 创建levelDB数据库
	dbFilePath := dbPath + nodeInfo
	err, db := OpenDatabase(dbFilePath, true)
	if err != nil {
		return nil, err
	}
	newBC.LevelDB = db
	// 2. 创建创世交易、区块
	genesisBlock := sideBlock.NewGenesisBlock(tx.NewCoinbaseTx(addr, "", int(configs.CoinBaseReward)), addr)
	// 3. 存储区块、最后Hash、高度到数据库
	batch := new(leveldb.Batch)
	batch.Put(append([]byte(configs.BlockPrefix), genesisBlock.BlockID...), genesisBlock.Serialization())
	batch.Put([]byte(configs.LatestHashKey), genesisBlock.BlockID)
	batch.Put([]byte(configs.BlockHeightKey),  []byte("0"))
	if err = db.Write(batch, nil); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// 4. 设置最新Hash
	newBC.LatestHash = genesisBlock.BlockID
	return &newBC, nil
}

// NewEmptySideBlockChain
// @Description: 创建一个空区块链（没有创世区块）
// @param dbPath
// @param nodeInfo
// @return *Blockchain
func NewEmptySideBlockChain(dbPath string, nodeInfo string) (*Blockchain, error) {
	var newBC Blockchain
	dbFilePath := dbPath + nodeInfo
	err, db := OpenDatabase(dbFilePath, true)
	if err != nil {
		return nil ,err
	}
	newBC.LevelDB = db
	newBC.LatestHash = tx.InitialHash
	if err = newBC.LevelDB.Put([]byte(configs.LatestHashKey), tx.InitialHash, nil); err != nil {
		logger.Logger.Error(err)
		return nil ,err
	}
	if err = newBC.LevelDB.Put([]byte(configs.BlockHeightKey), []byte("-1"), nil); err != nil {
		logger.Logger.Error(err)
		return nil ,err
	}
	return &newBC, nil
}

// LoadSideBlockChain
// @Description: 加载本地区块链
// @param nodeInfo	节点信息
// @return *Blockchain
func LoadSideBlockChain(dbPath string, nodeInfo string) *Blockchain {
	// 1. 判断区块链文件是否存在
	dbFilePath := dbPath + nodeInfo
	if !utils.DirectoryExists(dbFilePath) {
		logger.Logger.Error("No existing side-Blockchain found. Create one first.")
		os.Exit(1)
	}
	err, db := OpenDatabase(dbFilePath, false)
	// 2. 加载区块链
	latestHash, err := db.Get([]byte(configs.LatestHashKey), nil)
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	return &Blockchain{
		LatestHash: latestHash,
		LevelDB:    db,
	}
}

// StdOutputSideBlockChain
// @Description: 标准化输出区块链所有区块数据
// @receiver bc
func (bc *Blockchain) StdOutputSideBlockChain()  {
	iterator := bc.NewIterator()
	block := iterator.Next()
	for block != nil {
		fmt.Println(block)
		block = iterator.Next()
	}
}

// AddNewBlock
// @Description: 添加区块
// @receiver bc
// @param spvProof
// @param task
// @param epochIdx
// @param txs
// @return *sideBlock.SideBlock
func (bc *Blockchain) AddNewBlock(spvProof string, task task.Task, epochIdx int64, txs []*tx.Transaction) (*sideBlock.SideBlock, error) {
	newBlock := sideBlock.MiningBlock(spvProof, task, epochIdx, int64(bc.GetBestHeight() + 1), bc.LatestHash, txs)
	if _, err := bc.LinkNewBlock(newBlock); err != nil {
		return newBlock, err
	}
	return newBlock, nil
}

// NewIterator
// @Description: 创建迭代器
// @receiver bc
// @return *BlockChainIterator
func (bc *Blockchain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{
		currentHash: bc.LatestHash,
		db:          bc.LevelDB,
	}
}

// SignTransaction
// @Description: 签名当前交易
// @receiver bc
// @param tx
// @param priKey
func (bc *Blockchain) SignTransaction(tx *tx.Transaction, priKey ecdsa.PrivateKey)  {
	preTxs := bc.GetPreTx(tx)
	// 签名单个交易
	tx.Sign(preTxs, priKey)
}

// VerifyTx
// @Description: 验证所有交易
// @receiver bc
// @param tx
// @return bool
func (bc *Blockchain) VerifyTx(tx *tx.Transaction) bool {
	preTxs := bc.GetPreTx(tx)
	// 验证单个交易合法性
	return tx.Verify(preTxs)
}

// GetPreTx
// @Description: 创建前向交易集的映射关系，方便签名
// @receiver bc
// @param ntx
// @return map[string]*tx.Transaction
func (bc *Blockchain) GetPreTx(ntx *tx.Transaction) map[string]*tx.Transaction {
	// 如果当前交易是coinbase则不用查找前面的交易
	if ntx.IsCoinbase() {
		return nil
	}
	preTxs := make(map[string]*tx.Transaction)
	for _, input := range ntx.Inputs {
		// 查找
		ftx, err := bc.FindTransactionByTxHash(input.OutputTxHash)
		if err != nil {
			logger.Logger.Error(err)
			os.Exit(1)
		}
		// 加入集合
		preTxs[hex.EncodeToString(ftx.TxHash)] = ftx
	}
	// 返回
	return preTxs
}

// FindTransactionByTxHash
// @Description: 根据交易Hash查找对应的交易
// @receiver bc
// @param txHash
// @return *tx.Transaction
// @return error
func (bc *Blockchain) FindTransactionByTxHash(txHash []byte) (*tx.Transaction, error) {
	iterator := bc.NewIterator()
	block := iterator.Next()
	for block != nil {
		for _, tx := range block.Txs {
			if bytes.Compare(tx.TxHash, txHash) == 0 {
				return tx, nil
			}
		}
		block = iterator.Next()
	}
	return nil, errors.New(" Transaction Not Found!")
}

// GetBestHeight
// @Description: 获取当前区块链最大高度
// @receiver bc
// @return int
func (bc *Blockchain) GetBestHeight() int {
	heightByte, err := bc.LevelDB.Get([]byte(configs.BlockHeightKey), nil)
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	height, _ := strconv.Atoi(string(heightByte))
	return height
}

// GetBlockHashes
// @Description: 获取本地区块链所有Hash对照表
// @receiver bc
// @return map[int64][]byte
func (bc *Blockchain) GetBlockHashes() map[int64][]byte {
	hashes := make(map[int64][]byte)
	iterator := bc.NewIterator()
	block := iterator.Next()
	for block != nil {
		hashes[block.Height] = block.BlockID
		block = iterator.Next()
	}
	return hashes
}

// GetBlock
// @Description: 根据区块Hash获得区块
// @receiver bc
// @param hash
// @return *sideBlock.SideBlock
func (bc *Blockchain) GetBlock(hash []byte) *sideBlock.SideBlock {
	iterator := bc.NewIterator()
	block := iterator.Next()
	for block != nil {
		if bytes.Compare(block.BlockID, hash) == 0 {
			return block
		}
		block = iterator.Next()
	}
	return nil
}

// VerifyNewBlock
// @Description: 验证接收到的新区块
// @receiver bc
// @param newBlock
// @return bool
func (bc *Blockchain) VerifyNewBlock(newBlock *sideBlock.SideBlock) bool {
	// 1. 验证hash连续、高度
	if bytes.Compare(bc.LatestHash, newBlock.Previous) != 0 || newBlock.Height-1 != int64(bc.GetBestHeight()) {
		return false
	}
	// 2. 验证区块中交易合法性
	for _, btx := range newBlock.Txs {
		if !bc.VerifyTx(btx) {
			logger.Logger.Warn("VerifyTxERROR: Invalid transaction")
			return false
		}
	}
	// TODO 3. 验证PoFS共识难度
	return true
}

// LinkNewBlock
// @Description: 链接一个新区块
// @receiver bc
// @param newBlock
// @return int
// @return error
func (bc *Blockchain) LinkNewBlock(newBlock *sideBlock.SideBlock) (int, error) {
	if !bc.VerifyNewBlock(newBlock) {
		return -1, errors.New(" Invalid block, discarded.")
	}
	newHeight := bc.GetBestHeight() + 1
	batch := new(leveldb.Batch)
	batch.Put(append([]byte(configs.BlockPrefix), newBlock.BlockID...), newBlock.Serialization())
	batch.Put([]byte(configs.LatestHashKey), newBlock.BlockID)
	batch.Put([]byte(configs.BlockHeightKey), []byte(strconv.Itoa(newHeight)))
	if err := bc.LevelDB.Write(batch, nil); err != nil {
		return -1, err
	}
	bc.LatestHash = newBlock.BlockID
	return newHeight, nil
}