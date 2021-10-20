package tx

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fedSharing/sidechain/account"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/logger"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

var InitialHash = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

type Transaction struct {
	TxHash []byte
	Timestamp int64
	Inputs []*Input
	Outputs []*Output
}

// NewCoinbaseTx
// @Description: 	创世交易
// @param to 		接收者
// @param data		数据域
// @param reward	出块奖励
// @return *Transaction
func NewCoinbaseTx(to, data string, reward int) *Transaction {
	if data == "" {
		data = configs.GenesisCoinbaseData
	}
	// 创建输入与输出
	txInput := &Input{
		OutputTxHash: InitialHash,
		OutputIdx:    -1,				// 没有前置Hash
		Signature:    nil, 				// coinBase值
		PublicKey:    []byte(data),
	}
	txOutput := &Output{
		Value:        	reward,
		PublicKeyHash: 	account.ResolveAddressToPubKeyHash(to),
	}
	newTx :=  &Transaction{
		TxHash:    	  nil,
		Inputs:       []*Input{txInput},
		Outputs:      []*Output{txOutput},
		Timestamp: 	time.Now().UnixNano(),
	}
	newTx.SetHash()
	return newTx
}

// Serialize
// @Description: 交易结构体序列化
// @receiver t
// @return []byte
func (t *Transaction) Serialize() []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(t)
	if err != nil {
		log.Panic(err)
	}
	return buf.Bytes()
}

// DeserializeTx
// @Description: 交易结构体反序列化
// @param TxBytes
// @return *Transaction
func DeserializeTx(TxBytes []byte) *Transaction {
	var t Transaction
	decoder := gob.NewDecoder(bytes.NewReader(TxBytes))
	err := decoder.Decode(&t)
	if err != nil {
		log.Panic(err)
	}
	return &t
}

// SetHash
// @Description: 计算交易摘要
// @receiver t
func (t *Transaction) SetHash() {
	tBytes, _ := json.Marshal(t)
	res := sha256.Sum256(tBytes)
	t.TxHash = res[:]
}

// String
// @Description: 标准输出交易
// @receiver t
// @return string
func (t *Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("～～～～～～～～～～～～～～～～～～～～～～～～～～\n"))
	lines = append(lines, fmt.Sprintf("\tTxHash: %x\n", t.TxHash))
	lines = append(lines, fmt.Sprintf("\t\tInputs: \n"))
	for i, input := range t.Inputs {
		lines = append(lines, fmt.Sprintf("\t\t\t[%d]:\n", i))
		lines = append(lines, fmt.Sprintln(input.String()))
	}
	lines = append(lines, fmt.Sprintf("\t\tOutputs: \n"))
	for i, output := range t.Outputs {
		lines = append(lines, fmt.Sprintf("\t\t\t[%d]:\n", i))
		lines = append(lines, fmt.Sprintln(output.String()))
	}
	lines = append(lines, fmt.Sprintf("\t交易时间: %s\n", time.Unix(t.Timestamp/1e9 ,  0).Format("2006-01-02 15:04:05")))
	//lines = append(lines, fmt.Sprintf("\t交易时间: %s\n", strconv.FormatInt(t.Timestamp, 10)))
	lines = append(lines, fmt.Sprintf("～～～～～～～～～～～～～～～～～～～～～～～～～～\n"))
	return strings.Join(lines, "")
}

// IsCoinbase
// @Description: 验证一个交易为创世交易
// @receiver t
// @return bool
func (t *Transaction) IsCoinbase() bool {
	return len(t.Inputs) == 1 && string(t.Inputs[0].OutputTxHash) == string(InitialHash) && t.Inputs[0].OutputIdx == -1
}

// Sign
// @Description: 对于每个交易的输出签名
// @receiver t
// @param preTxs
// @param priKey
func (t *Transaction) Sign(preTxs map[string]*Transaction, priKey ecdsa.PrivateKey)  {
	// coinbase交易没有输入不签名
	if t.IsCoinbase() {
		return
	}
	// 验证此交易vin的所有前置交易都不为空
	for _, vin := range t.Inputs {
		if preTxs[hex.EncodeToString(vin.OutputTxHash)].TxHash == nil {
			logger.Logger.Error("ERROR: Previous transaction is not correct")
			os.Exit(1)
		}
	}
	// 复制一个裁剪过的交易
	txCopy := t.TrimmedCopy()
	// 签名该交易所有的Input,默认一个交易的输入只有一个
	for inputID, vin := range txCopy.Inputs {
		// 获取当前vin的前置交易
		prevTx := preTxs[hex.EncodeToString(vin.OutputTxHash)]
		// 获取PubKeyHash并赋值到vin的Pubkey上（临时存储，为了取Hash）
		txCopy.Inputs[inputID].Signature = nil
		txCopy.Inputs[inputID].PublicKey = prevTx.Outputs[vin.OutputIdx].PublicKeyHash
		// 交易取Hash,这样就包含了出款人的pubHashKey（在input的pubkey中）和收款人的pubHashKey（在此交易的vout中）
		txCopy.SetHash()
		// 设置回nil
		txCopy.Inputs[inputID].PublicKey = nil

		// 对此交易整体进行签名(使用私钥)
		r, s, err := ecdsa.Sign(rand.Reader, &priKey, txCopy.TxHash)
		if err != nil {
			logger.Logger.Error(err)
			os.Exit(1)
		}
		// 签名(也就是r和s的字节组合)  注意这里赋值给t，而不是txcopy
		t.Inputs[inputID].Signature = append(r.Bytes(), s.Bytes()...)
	}
}

// Verify
// @Description: 验证交易签名
// @receiver t
// @param preTxs
// @return bool
func (t *Transaction) Verify(preTxs map[string]*Transaction) bool {
	if t.IsCoinbase() {
		return true
	}
	for _, vin := range t.Inputs {
		if preTxs[hex.EncodeToString(vin.OutputTxHash)].TxHash == nil {
			logger.Logger.Error("ERROR: Previous transaction is not correct")
			os.Exit(1)
		}
	}
	txCopy := t.TrimmedCopy()
	curve := elliptic.P256()
	// 验证交易中的每个Input
	for inputID, vin := range t.Inputs {
		// 获取当前vin的前置交易
		prevTx := preTxs[hex.EncodeToString(vin.OutputTxHash)]
		// 获取PubKeyHash并赋值到vin的Pubkey上（临时存储，为了取Hash）
		txCopy.Inputs[inputID].Signature = nil
		txCopy.Inputs[inputID].PublicKey = prevTx.Outputs[vin.OutputIdx].PublicKeyHash
		// 交易取Hash,这样就包含了出款人的pubHashKey（在input的pubkey中）和收款人的pubHashKey（在此交易的vout中）
		txCopy.SetHash()
		// 设置回nil
		txCopy.Inputs[inputID].PublicKey = nil
		// 验证签名
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen)/2])	//r前半段
		s.SetBytes(vin.Signature[(sigLen)/2:])	//s后半段

		// 分割vin的公钥
		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.PublicKey[(keyLen / 2):])


		// 验证
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.TxHash, &r, &s) {
			return false
		}
	}
	return true
}

// TrimmedCopy
// @Description: 裁剪一个交易并拷贝(为了构造签名结构体)
// @receiver t
// @return *Transaction
func (t *Transaction) TrimmedCopy() *Transaction {
	var inputs []*Input
	var outputs []*Output
	for _, vin := range t.Inputs {
		// 注意就是将签名和公钥字段设置为nil
		inputs = append(inputs, &Input{vin.OutputTxHash, vin.OutputIdx, nil, nil})
	}
	for _, vout := range t.Outputs {
		outputs = append(outputs, &Output{vout.Value, vout.PublicKeyHash})
	}
	// 拷贝
	txCopy := &Transaction{
		TxHash:    t.TxHash,
		Timestamp: t.Timestamp,
		Inputs:    inputs,
		Outputs:   outputs,
	}

	return txCopy
}