package account

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fedSharing/sidechain/logger"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Wallet struct {
	KeyPairs map[string]*KeyPair
}

// InitRandomWallet
// @Description: 初始化一个本地钱包(随机10个账户)
// @return *Wallet
func InitRandomWallet() *Wallet {
	var w Wallet
	w.KeyPairs = make(map[string]*KeyPair)
	for i := 0; i < 10; i++ {
		w.AddNewKeyPair()
	}
	return &w
}

// LoadLocalWalletFile
// @Description: 加载本地钱包文件，获得所有账户
// @return *Wallet
func LoadLocalWalletFile(path string) *Wallet {
	var w Wallet
	w.KeyPairs = make(map[string]*KeyPair)
	w.LoadFromFile(path)
	return &w
}

// LoadFromFile
// @Description: 加载本地文件
// @receiver w
func (w *Wallet) LoadFromFile(path string) {
	// 序列化存储为一个dat文件
	// 1. 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Logger.Error("There is no wallet file, Please create first!")
		os.Exit(1)
	}
	// 2. 读取文件
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	// 3. 解码/反序列化
	var wallet Wallet
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileBytes))
	// 4. 加载到wallets结构体
	err = decoder.Decode(&wallet)
	if err != nil {
		log.Panic(err)
	}
	// 5. 赋值
	w.KeyPairs = wallet.KeyPairs
}

// SaveToFile
// @Description: 将钱包永久化存储到本地文件中
// @param path: 路径
// @receiver w
func (w *Wallet) SaveToFile(path string)  {
	// 1. 序列化
	content := new(bytes.Buffer)       // 创建缓冲区
	gob.Register(elliptic.P256())      // 注册
	encoder := gob.NewEncoder(content) // 创建encoder
	err := encoder.Encode(w)          // 序列化wallets
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	// 2. 打开文件,写入
	if err = ioutil.WriteFile(path, content.Bytes(), 0644); err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
}

// AddNewKeyPair
// @Description: 在钱包中添加一个账户(一对公私钥)
// @receiver w
// @return string
func (w *Wallet) AddNewKeyPair() string {
	kp := NewKeyPairStruct()
	address := kp.GetAddress()
	w.KeyPairs[string(address)] = kp
	return string(address)
}

// GetAddresses
// @Description: 获取本地钱包的所有地址
// @receiver w
// @return addresses
func (w *Wallet) GetAddresses() (addresses []string) {
	for k := range w.KeyPairs {
		addresses = append(addresses, k)
	}
	return
}

func (w *Wallet) String() string {
	var line []string
	line = append(line, fmt.Sprintf("\n-------------------------------------------------\n"))
	i := 0
	for addr := range w.KeyPairs {
		line = append(line, fmt.Sprintf("[%d]: addr = %s\n", i, addr))
		i ++
	}
	line = append(line, fmt.Sprintf("-------------------------------------------------\n"))
	return strings.Join(line, "")
}