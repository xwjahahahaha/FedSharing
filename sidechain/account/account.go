package account

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fedSharing/sidechain/configs"
	"fedSharing/sidechain/logger"
	"fedSharing/sidechain/utils"
	"fedSharing/sidechain/utils/crypto"
	"golang.org/x/crypto/ripemd160"
	"os"
)

type KeyPair struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey []byte
}

// NewKeyPairStruct
// @Description: 创建密钥对
// @return *KeyPair
func NewKeyPairStruct() *KeyPair {
	privateKey, publicKey := NewKeyPair()
	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}

// NewKeyPair
// @Description: 使用椭圆曲线创建新的密钥对
// @return ecdsa.PrivateKey
// @return []byte
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()									// 创建曲线
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)	// 创建私钥
	if err != nil {
		logger.Logger.Error("create key pair err: ", err)
		os.Exit(1)
	}
	// 创建公钥
	// 基于椭圆曲线，公钥是曲线上的点，所以公钥是X，Y坐标的组合, 在比特币中将其连接起来实现公钥
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return *privateKey, publicKey
}

// GetAddress
// @Description: 公钥转换为地址
// @receiver kp
// @return []byte
func (kp *KeyPair) GetAddress() []byte {
	publicSHA256 := HashPubKey(kp.PublicKey)
	versionByte := utils.Int64ToBytes(int64(configs.PubKeyVersion))
	versionedPayload := append(versionByte[len(versionByte)-1:], publicSHA256...)
	checkSum := checksum(versionedPayload)
	fullPayload := append(versionedPayload, checkSum...)
	address := crypto.Base58Encode(fullPayload)
	return append([]byte(configs.AddressPrefix), address...)
}

// HashPubKey
// @Description: 对公钥先SHA256再RIPEMD160加密
// @param pubKey
// @return []byte
func HashPubKey(pubKey []byte) []byte {
	// 先sha256再RIPEMD160加密
	pubKeySha256 := sha256.Sum256(pubKey)
	// 创建RIPEMD160
	// 注意这里用golang.org/x/crypto/ripemd160包下的ripemd160
	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(pubKeySha256[:])
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	return RIPEMD160Hasher.Sum(nil)
}

// checksum
// @Description: 计算校验和 sha256(sha256(PublicKey))[:4]
// @param payload
// @return []byte
func checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	// 截取hash后AddressCheckSumLen个字节
	return second[:configs.AddressCheckSumLen]
}

// ValidateAddress
// @Description: 验证地址有效性
// @param address
// @return bool
func ValidateAddress(address string) bool  {
	if len(address) < 37 {
		logger.Logger.Error("Too short address!")
		return false
	}
	address = address[len([]byte(configs.AddressPrefix)):]
	// 1. base58解码
	pubKeyHash := crypto.Base58Decode([]byte(address))
	// 2. 获取checkSum和pubHashKey部分
	// prefixLoad [1:] 是去掉Base58 Decode前面添加的一个字节0
	prefixLoad, actualChecksum := pubKeyHash[1:len(pubKeyHash)-configs.AddressCheckSumLen], pubKeyHash[len(pubKeyHash)-configs.AddressCheckSumLen:]
	// 3. 计算sum
	nowChecksum := checksum(prefixLoad)
	return bytes.Compare(nowChecksum, actualChecksum) == 0
}

// ResolveAddressToPubKeyHash
// @Description: 解析地址为PubKeyHash
// @param address
// @return []byte
func ResolveAddressToPubKeyHash(address string) []byte {
	if !ValidateAddress(address) {
		logger.Logger.Error("InValid Address!")
		os.Exit(1)
	}
	address = address[len([]byte(configs.AddressPrefix)):]
	// base58解码
	pubKeyHash := crypto.Base58Decode([]byte(address))
	// 2byte 是因为 1byte由Decode造成前面1byte的0，剩下1byte是version
	return pubKeyHash[2:len(pubKeyHash)-configs.AddressCheckSumLen]
}

