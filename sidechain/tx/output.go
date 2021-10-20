package tx

import (
	"bytes"
	"encoding/gob"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/utils/crypto"
	"fmt"
	"log"
	"strings"
)

type Output struct {
	Value int
	PublicKeyHash []byte
}

type Outputs struct {
	Outputs []*Output
}

// String
// @Description: 标准输出Output
// @receiver out
// @return string
func (out *Output) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("\t\t\t\tValue : %d\n", out.Value))
	lines = append(lines, fmt.Sprintf("\t\t\t\tPubKeyHash : %x\n", out.PublicKeyHash))
	return strings.Join(lines, "")
}

// Serialize
// @Description: 序列化输出数组
// @receiver out
// @return []byte
func (outs *Outputs) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return res.Bytes()
}

// DeSerializeOuts
// @Description: 反序列化输出
// @param outsBytes
// @return *Outputs
func DeSerializeOuts(outsBytes []byte) *Outputs {
	var outs Outputs
	decoder := gob.NewDecoder(bytes.NewReader(outsBytes))
	err := decoder.Decode(&outs)
	if err != nil {
		log.Panic(err)
	}
	return &outs
}

// Lock
// @Description: 对于当前输出设置其PublicKeyHash(上锁)
// @receiver out
// @param address
func (out *Output) Lock(address []byte){
	address = address[len([]byte(configs.AddressPrefix)):]
	// 1. 将地址解码base58
	pubKeyHash := crypto.Base58Decode(address)
	// 2. 截取中间段就是PubKeyHash
	// 前一个是0，后一个byte是version，后四个是checksum
	pubKeyHash = pubKeyHash[2:len(pubKeyHash)-configs.AddressCheckSumLen]
	// 3. 设置
	out.PublicKeyHash = pubKeyHash
}

// IsLockedWithKey
// @Description: 检查此公钥hash与输出的PublicKeyHash是否相同
// @receiver out
// @param pubKeyHash
// @return bool
func (out *Output) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(pubKeyHash, out.PublicKeyHash) == 0
}

// NewTxOutput
// @Description: 创建一个新的交易输出
// @param value
// @param address
// @return *Output
func NewTxOutput(value int, address string) *Output {
	newOutput :=  &Output{
		Value:      value,
		PublicKeyHash: nil,
	}
	newOutput.Lock([]byte(address)) 		// 上锁（地址 => PubkeyHash）
	return newOutput
}
