package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strconv"
)

const CommandLength = 12

// Int64ToBytes
// @Description: Int64转换为byte数组
// @param num
// @return []byte
func Int64ToBytes(num int64) []byte {
	buff := new(bytes.Buffer)
	//BigEndian指定大端小端
	//binary.Write是将数据的二进制格式写入字节缓冲区中
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// BytesToInt
// @Description: byte数组转换为int64
// @param b
// @return int64
func BytesToInt(b []byte) int64 {
	bytesBuffer := bytes.NewBuffer(b)
	var x int64
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

// AssemblySocketAddr
// @Description: 组装完整的socket地址
// @param ip
// @param port
// @return string
func AssemblySocketAddr(ip string, port int) string {
	return ip + ":" + strconv.Itoa(port)
}


// GobEncode
// @Description: 序列化结构体
// @param data
// @return []byte
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

// DirectoryExists
// @Description: 判断目录是否存在
// @param path
// @return bool
func DirectoryExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}else if os.IsNotExist(err) {
		return false
	}
	return false
}