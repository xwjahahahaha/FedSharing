package merkletree

import (
	"crypto/sha256"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/logger"
	"math"
)

type MerkleTree struct {
	Nodes [][]*MerkleNode
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left *MerkleNode
	Right *MerkleNode
	Data []byte
}

var FillData = []byte{0}		// 完美二叉树填充数据

// NewMerkleNode
// @Description: 创建一个新的Merkle节点
// @param left	左孩子
// @param right	右孩子
// @param data
// @return *MerkleNode
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}
	// 计算当前节点数据域
	if left == nil && right == nil {		// 叶子节点
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	}else {									// 上层节点
		preHash := append(left.Data, right.Data...)
		hash := sha256.Sum256(preHash)
		mNode.Data = hash[:]
	}
	mNode.Left = left
	mNode.Right = right
	return &mNode
}

// NewMerkleTree
// @Description: 创建 Merkle tree
// @param data
// @return *MerkleTree
func NewMerkleTree(data [][]byte) *MerkleTree {
	n := len(data)
	if n == 0 {
		logger.Logger.Warn("NewMerkleTree: There is not have txs in merkletree!")
		return &MerkleTree{RootNode: NewMerkleNode(nil, nil, FillData)}
	}else if n > configs.MaxTxsAmount {
		logger.Logger.Warn("NewMerkleTree: Beyond the maximum number of tx in a block")
		data = data[:configs.MaxTxsAmount]
		logger.Logger.Warn("NewMerkleTree: Let go of the subsequent excess transactions")
	}
	// 确保完美二叉树，总数必须为2的倍数
	if gap := IsMultipleOfTwo(n) - n; gap != 0 {						// TODO 冗余可优化
		for i := 0; i < gap; i++ {
			data = append(data, FillData)
		}
	}
	allLevel := int(math.Log2(float64(len(data)))) + 1
	// 新建MerkleTree
	newMerkleTree, level := &MerkleTree{
		Nodes:    make([][]*MerkleNode, allLevel),
		RootNode: nil,
	}, 0
	// 生成交易节点/叶子节点(第一层)
	for _, txData := range data {
		newMerkleTree.Nodes[level] = append(newMerkleTree.Nodes[level], NewMerkleNode(nil, nil, txData))
	}
	// 依层建立节点，总层数log2^(len(data)) (除去底层的交易节点/叶子节点的一层)
	for level++; level < allLevel; level++ {
		oldLevel, newLevel := newMerkleTree.Nodes[level-1], make([]*MerkleNode, 0)
		// 遍历节点
		for j := 0; j < len(oldLevel); j += 2 {			// 两两选择计算新的上层节点
			newLevel = append(newLevel, NewMerkleNode(oldLevel[j], oldLevel[j+1], nil))
		}
		// 更新新一层的节点
		newMerkleTree.Nodes[level] = newLevel
	}
	// 构建完毕，最后一层的第一个节点就是merkle根
	newMerkleTree.RootNode = newMerkleTree.Nodes[level-1][0]
	return newMerkleTree
}

// IsMultipleOfTwo
// @Description: 寻找最贴近的2的倍数
// @param num
// @return multiple
func IsMultipleOfTwo(num int) int {
	// 找到一个2的倍数 >= 该数量
	if num <= 0 {
		return 0
	}else if num == 1 {
		return 2
	}
	multiple := 1
	for ; multiple < num; multiple *= 2 {}
	return multiple
}