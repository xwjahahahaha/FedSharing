package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/logger"
	"fedSharing/sidechain/merkletree"
	"fedSharing/sidechain/pofs"
	"fedSharing/sidechain/task"
	"fedSharing/sidechain/tx"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"
)

var (
	Miner = "xwj"		//TODO
)

type SideBlock struct {
	BlockID []byte
	Previous []byte
	Height int64
	Miner []byte
	AvgVarphi map[string]*big.Float
	Varphi map[string]*big.Float
	Difficulty float64
	TaskID int64
	EpochID int64
	MerkleTreeRoot []byte
	MerkleTree *merkletree.MerkleTree
	Timestamp int64

	Txs []*tx.Transaction
}

// NewGenesisBlock
// @Description: 创世区块(生成新的区块链时创建)
// @param coinbaseTx
// @return *SideBlock
func NewGenesisBlock(coinbaseTx *tx.Transaction, miner string) *SideBlock {
	genesisBlock := &SideBlock{
		BlockID:        nil,
		Previous:       tx.InitialHash,
		Height:         0,
		Miner:          []byte(miner),
		AvgVarphi:      nil,
		Varphi:         nil,
		Difficulty:     configs.Difficulty,
		TaskID:         -1,
		EpochID:        -1,
		MerkleTreeRoot: nil,
		MerkleTree: 	nil,
		Timestamp:      time.Now().Unix(),
		Txs:            []*tx.Transaction{coinbaseTx},
	}
	// 设置MerkleTree
	nmt := merkletree.NewMerkleTree(genesisBlock.SerializeTxs())
	genesisBlock.MerkleTreeRoot = nmt.RootNode.Data
	genesisBlock.MerkleTree = nmt
	// 设置区块Hash
	genesisBlock.SetHash()
	return genesisBlock
}

// MiningBlock
// @Description: 创建新区块
// @param spvProof
// @param task
// @param epochIdx
// @param height
// @param preHash
// @return *SideBlock
func MiningBlock(spvProof string, task task.Task, epochIdx int64, height int64, preHash []byte, txs []*tx.Transaction) *SideBlock {
	// 1. 验证spv
	valid, _ := VerifyDeposit(spvProof)
	if !valid {
		logger.Logger.Warn("请先在主链上质押代币并提交正确的主链SPV证明！")
		return nil
	}
	// 2.TODO 验证所有交易

	// TODO 远程调用Pool Manager获取本轮必要参数 通过 taskID, epochIdx

	// 3. 运行PoFS共识
	pofs := pofs.NewPoFS(task, epochIdx)
	local_varphi := pofs.Run()
	// 4. 广播自己运算结果

	// 5. 收集一定数量的结果，数量为超过节点总数的2/3,计算平均值

	// 6. 计算与本地距离，比较难度

	// 7. 符合后计算Hash出块
	newBlock := &SideBlock{
		BlockID:        nil,
		Previous:       preHash,
		Height:         height,
		Miner:          []byte(Miner),		//TODO
		AvgVarphi:      nil,				//TODO
		Varphi:         local_varphi,
		EpochID:	 	epochIdx,
		Difficulty:     0,
		TaskID:         task.Idx,
		MerkleTreeRoot: nil,		//TODO
		Timestamp:      time.Now().Unix(),
		Txs:            txs,
	}
	newBlock.SetHash()
	return newBlock
}

// TODO 验证主链质押,并按比例转换出块奖励
func VerifyDeposit(spvProof string) (bool, int) {

	return true, 10
}

func (b *SideBlock) SetHash()  {
	sum256 := sha256.Sum256(b.Serialization())
	b.BlockID = sum256[:]
}

// Serialization
// @Description: 区块序列化
// @receiver b
// @return []byte
func (b *SideBlock) Serialization() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	if err != nil {
		logger.Logger.Error("Serialization err : ", err)
		return nil
	}
	return res.Bytes()
}

// DeserializationBlock
// @Description: 区块反序列化
// @param blockBytes
// @return *SideBlock
func DeserializationBlock(blockBytes []byte) *SideBlock {
	var block SideBlock
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&block)
	if err != nil {
		logger.Logger.Error("DeserializationBlock err : ", err)
		os.Exit(1)
	}
	return &block
}

// String
// @Description: 标准输出区块
// @receiver b
// @return string
func (b *SideBlock) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("==================================================================================================================\n"))
	lines = append(lines, fmt.Sprintf("区块高度: %d\n", b.Height))
	lines = append(lines, fmt.Sprintf("前区块hash值: %x\n", b.Previous))
	lines = append(lines, fmt.Sprintf("本区块Hash值: %x\n", b.BlockID))
	lines = append(lines, fmt.Sprintf("出块人: %s\n", b.Miner))
	lines = append(lines, fmt.Sprintf("Task编号: %d\n", b.TaskID))
	lines = append(lines, fmt.Sprintf("全局迭代轮次: %d\n", b.EpochID))
	timeFormat := time.Unix(b.Timestamp, 0).Format("2006-01-02 15:04:05")
	lines = append(lines, fmt.Sprintf("生成时间: %s\n", timeFormat))
	lines = append(lines, fmt.Sprintf("Shapley Value向量: \n"))
	for k ,v := range b.Varphi {
		lines = append(lines, fmt.Sprintf("\t[Member: %s, Varphi: %s]\n", k, v.Text('f', -1)))
	}
	lines = append(lines, "交易列表如下: \n")
	for i, tx := range b.Txs {
		lines = append(lines, fmt.Sprintf("[%d]:\n", i))
		lines = append(lines, fmt.Sprintf(tx.String()))
	}
	lines = append(lines, fmt.Sprintf("==================================================================================================================\n"))
	return strings.Join(lines, "")
}

// SerializeTxs
// @Description: 序列化所有交易
// @receiver b
// @return txBytes
func (b *SideBlock) SerializeTxs() (txBytes [][]byte) {
	for _, tx := range b.Txs {
		txBytes = append(txBytes, tx.Serialize())
	}
	return
}