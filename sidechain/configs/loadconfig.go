package configs

import (
	"gopkg.in/ini.v1"
)

var (
	GenesisCoinbaseData string

	Difficulty float64

	RootCmd string
	RootShort string
	RootLong string

	CoinBaseReward uint

	PubKeyVersion int			// 公钥版本
	AddressCheckSumLen int  	// 截取校验和Hash后字节数

	MaxTxsAmount int			// 区块中交易的最大数量

	ChainVersion int			// 侧链版本
)

func LoadBlockChainConfig(file *ini.File)  {
	GenesisCoinbaseData = file.Section("genesis").Key("GenesisCoinbaseData").MustString("The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")
}

func LoadConsensus(file *ini.File)  {
	Difficulty = file.Section("pofs").Key("Difficulty").MustFloat64(0.05)
}

func LoadDataBase(file *ini.File)  {
}

func LoadCmdConfig(file *ini.File)  {
	RootCmd = file.Section("cmd").Key("RootCmd").MustString("btc")
	RootShort = file.Section("cmd").Key("RootShort").MustString("Simple bitcoin by gump")
	RootLong = file.Section("cmd").Key("RootLong").MustString("Simple bitcoin by gump")
}

func LoadTransaction(file *ini.File)  {
	CoinBaseReward = file.Section("tx").Key("CoinBaseReward").MustUint(10)
}

func LoadEnCrypto(file *ini.File)  {
	PubKeyVersion = file.Section("encrypt").Key("PubKeyVersion").MustInt(1)
	AddressCheckSumLen = file.Section("encrypt").Key("AddressCheckSumLen").MustInt(4)
}

func loadServer(file *ini.File)  {

}

func LoadBlock(file *ini.File)  {
	MaxTxsAmount = file.Section("block").Key("MaxTxsAmount").MustInt(10)
}

func LoadBlockchain(file *ini.File)  {
	ChainVersion = file.Section("blockchain").Key("ChainVersion").MustInt(1)
}