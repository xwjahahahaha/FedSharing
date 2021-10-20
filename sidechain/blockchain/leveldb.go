package blockchain

import (
	"fedSharing/sidechain/logger"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
)

// OpenDatabase
// @Description: 打开或创建一个LevelDB数据库(记得关闭)
// @param path
// @exist	数据库已存在是否报错
// @return *leveldb.DB
func OpenDatabase(path string, exist bool) (error, *leveldb.DB) {
	db, err := leveldb.OpenFile(path, &opt.Options{
		ErrorIfExist: exist,
	})		// levelDB并发安全，支持并发操作
	if err != nil {
		if err == os.ErrExist {
			logger.Logger.Error("Side-Blockchain already exists.")
			return err, nil
		} else {
			logger.Logger.Error("区块链数据库开启失败！", err)
			os.Exit(1)
		}
	}
	if exist {
		logger.Logger.Info("区块链数据库创建成功！")
	}
	return nil, db
}
