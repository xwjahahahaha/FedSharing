package utils

import (
	"bytes"
	"encoding/gob"
	flog "fedSharing/mainchain/log"
	"fmt"
	"log"
	"os"
)

const CommandLength = 40

func DirectoryOrFileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}else if os.IsNotExist(err) {
		return false
	}
	return false
}

func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// CommandToBytes
// @Description: 命令转[]byte
// @param command
// @return []byte
func CommandToBytes(command string) []byte {
	var bytes [CommandLength]byte // 创建字节缓冲区

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

// BytesToCommand
// @Description: 字节数组转换为命令
// @param bytes
// @return string
func BytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

func ColorPrint(s string)  {
	flog.Logger.Infof("\x1b[32m%s\x1b[0m", s)
}