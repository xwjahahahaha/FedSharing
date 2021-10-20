package utxo

import (
	"errors"
	"fedSharing/sidechain/account"
	"fedSharing/sidechain/block"
	"fedSharing/sidechain/blockchain"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/logger"
	"fedSharing/sidechain/tx"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os"
	"time"
)

type UnSpendTxOutputSet struct {
	SBC *blockchain.Blockchain
}

// NewUTXO
// @Description: 新建一个UTXO模型
// @param bc
// @return *UnSpendTxOutputSet
func NewUTXO(bc *blockchain.Blockchain) *UnSpendTxOutputSet {
	return &UnSpendTxOutputSet{bc}
}

// FindAllUTXO
// @Description: 找到所有的UTXO
// @receiver u
// @return map[string][]*tx.Output
func (u *UnSpendTxOutputSet) FindAllUTXO() map[string]tx.Outputs {
	utxos := make(map[string]tx.Outputs)
	// 已访问标记数组
	visitedMap := make(map[string]map[int]bool, 0)
	// 遍历区块
	bci := u.SBC.NewIterator()
	block := bci.Next()
	for block != nil {
		// 遍历交易
		for _, btx := range block.Txs {
			outs := tx.Outputs{}
			for outIdx, out := range btx.Outputs {
				// 判断是否已被标记，是则表示Output已使用,直接跳过
				if _, has := visitedMap[string(btx.TxHash)][outIdx]; has {
					continue
				}
				// 加入到Utxo中
				outs.Outputs = append(outs.Outputs, out)
			}
			utxos[string(btx.TxHash)] = outs
			// 遍历input
			if btx.IsCoinbase() {
				// 排除coinbaseTx 前面没有输出
				continue
			}
			for _, input := range btx.Inputs {
				if outputMap, has := visitedMap[string(input.OutputTxHash)]; has{
					outputMap[input.OutputIdx] = true		// 标记
				}else{
					newMap := make(map[int]bool)
					newMap[input.OutputIdx] = true
					visitedMap[string(input.OutputTxHash)] = newMap
				}
			}
		}
		block = bci.Next()
	}
	return utxos
}

// Reindex
// @Description: 重新索引所有utxo
// @receiver u
func (u *UnSpendTxOutputSet) Reindex()  {
	// 删除之前所有的utxo
	batch := new(leveldb.Batch)
	iter := u.SBC.LevelDB.NewIterator(util.BytesPrefix([]byte(configs.TxPrefix)), nil)
	defer iter.Release()
	for iter.Next() {
		batch.Delete(iter.Key())
	}
	// 1. 获取当前区块链所有utxo
	utxo := u.FindAllUTXO()
	// 2. 批量插入
	for txHash, outs := range utxo {
		batch.Put([]byte(configs.TxPrefix + txHash), outs.Serialize())
	}
	if err := u.SBC.LevelDB.Write(batch, nil); err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
}

// FindSpendableOutputs
// @Description: 查询utxo池中满足某个人需求金额的所有utxo（用于构造交易）
// @receiver u
// @param publicKeyHash 用户公钥Hash
// @param amount	 	需要支付金额
// @return int
// @return map[string][]int
func (u *UnSpendTxOutputSet) FindSpendableOutputs(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	iter := u.SBC.LevelDB.NewIterator(util.BytesPrefix([]byte(configs.TxPrefix)), nil)
	defer iter.Release()
	for iter.Next() {
		txHash := string(iter.Key()[len(configs.TxPrefix):])
		outs := tx.DeSerializeOuts(iter.Value())
		for outIdx, out := range outs.Outputs {
			// 如果能够用对应的公钥解锁output 且累计金额还未到要求的金额，那么就继续累计utxo
			if out.IsLockedWithKey(publicKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txHash] = append(unspentOutputs[txHash], outIdx)
			}
		}
	}
	err := iter.Error()
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	return accumulated, unspentOutputs
}

// FindUTXOByOne
// @Description: 在UTXO池中查询一个人的所有UTXO
// @receiver u
// @param publicKeyHash	用户公钥Hash
// @return *tx.Outputs	output数组
func (u *UnSpendTxOutputSet) FindUTXOByOne(publicKeyHash []byte) *tx.Outputs {
	var utxo tx.Outputs
	iter := u.SBC.LevelDB.NewIterator(util.BytesPrefix([]byte(configs.TxPrefix)), nil)
	defer iter.Release()
	for iter.Next() {
		outs := tx.DeSerializeOuts(iter.Value())
		for _, out := range outs.Outputs {
			// 判断是否为本人
			if out.IsLockedWithKey(publicKeyHash) {
				utxo.Outputs = append(utxo.Outputs, out)
			}
		}
	}
	return &utxo
}


// NewTransaction
// @Description: 创建一个新的交易
// @param from		发起人
// @param to		接收人
// @param amount	金额
// @param utxo		utxo集合
// @return *Transaction
func (u *UnSpendTxOutputSet) NewTransaction(from, to string, amount int, walletPath string) (*tx.Transaction, error) {
	var inputs []*tx.Input
	var outputs []*tx.Output
	if len(from)==0 || len(to) == 0 || amount < 0 {
		logger.Logger.Error("The wrong input！")
		os.Exit(1)
	}
	// 加载本地钱包文件
	w := account.LoadLocalWalletFile(walletPath)
	// 获取from的密钥对
	fromKeyPair := w.KeyPairs[from]
	// 计算公钥的Hash
	publicKeyHash := account.HashPubKey(fromKeyPair.PublicKey)
	// 获取当前from可以支付的金额
	acc, outputsDesc := u.FindSpendableOutputs(publicKeyHash, amount)
	if acc < amount {
		logger.Logger.Info("Sorry, Your balance is insufficient.")
		return nil, errors.New(" balance is insufficient")
	}
	// 创建交易Input
	for txHash, outputIdxAry := range outputsDesc {
		for _, outputIdx := range outputIdxAry {
			inputs = append(inputs, &tx.Input{
				OutputTxHash: []byte(txHash),
				OutputIdx:    outputIdx,
				Signature:    nil,
				PublicKey:    fromKeyPair.PublicKey,	// 设置input中from的公钥
			})
		}
	}
	// 创建交易Output
	outputs = append(outputs, tx.NewTxOutput(amount, to))
	// 找零(如果有的话)
	if acc > amount {
		outputs = append(outputs, tx.NewTxOutput(acc - amount, from))
	}
	ntx := &tx.Transaction{
		TxHash:    	  nil,
		Inputs:       inputs,
		Outputs:      outputs,
		Timestamp: time.Now().UnixNano(),
	}
	ntx.SetHash()
	u.SBC.SignTransaction(ntx, fromKeyPair.PrivateKey)
	return ntx, nil
}

func (u *UnSpendTxOutputSet) Update(block *block.SideBlock)  {
	db := u.SBC.LevelDB
	batch := new(leveldb.Batch)
	for _, btx := range block.Txs {
		if !btx.IsCoinbase() {
			// 非coinbase交易处理
			for _, input := range btx.Inputs {
				// 遍历交易的前置关联交易中的输出，做更新：被引用->删除，否则不变
				previousOutsBytes, err := db.Get(append([]byte(configs.TxPrefix), input.OutputTxHash...), nil)
				if err != nil {
					logger.Logger.Error(fmt.Sprintf("Con't find %x in	this Blockchain Database.", input.OutputTxHash))
					os.Exit(1)
				}
				previousOuts := tx.DeSerializeOuts(previousOutsBytes)
				updateOuts := tx.Outputs{}
				for outputIdx, output := range previousOuts.Outputs {
					// 以选择加入代替删除操作
					if input.OutputIdx != outputIdx {
						// 不等于，那么说明没有被新区块中的input引用，所以加入
						updateOuts.Outputs = append(updateOuts.Outputs, output)
					}
				}
				if len(updateOuts.Outputs) == 0 {
					// 都被删完了,直接在数据库中删除掉这个交易
					batch.Delete(append([]byte(configs.TxPrefix), input.OutputTxHash...))
				}else {
					// 否则更新
					batch.Put(append([]byte(configs.TxPrefix), input.OutputTxHash...), updateOuts.Serialize())
				}
			}
		} else {
			// coinbase交易处理，无需标记直接加入UTXO池
			outputs := tx.Outputs{}
			for _, output := range btx.Outputs {
				outputs.Outputs = append(outputs.Outputs, output)
			}
			batch.Put(append([]byte(configs.TxPrefix), btx.TxHash...), outputs.Serialize())
		}
	}
	if err := db.Write(batch, nil); err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
}