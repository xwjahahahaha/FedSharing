package node

import (
	"bufio"
	"bytes"
	"fedSharing/mainchain/configs"
	"fedSharing/mainchain/log"
	"fedSharing/mainchain/measure"
	"fedSharing/mainchain/utils"
	"github.com/libp2p/go-libp2p-core/network"
	"io"
	"os"
	"strconv"
	"time"
)

const StreamFileDeliveryFlag = "@fileOver"

func handleDiffStream(s network.Stream) {
	// 根据新的流创建读写流
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go writeDiffData(rw, s)
	// 流s始终开启，直到流两端的任何一方关闭他
}

func handleModelStream(s network.Stream) {
	// 根据新的流创建读写流
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go ReadModelData(rw, s)

	// 流s始终开启，直到流两端的任何一方关闭他
}

// readDiffData
// @Description: 读取Diff文件
// @param rw
func readDiffData(rw *bufio.ReadWriter, diffFilePath string) {
	file, err := os.OpenFile(diffFilePath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Logger.Error(err)
		return
	}
	writer := bufio.NewWriter(file)
	buf := make([]byte, 1024*4)
	for {
		n, err := rw.Read(buf)
		if err != nil {
			log.Logger.Error(err)
			return
		}
		if bytes.Compare([]byte(StreamFileDeliveryFlag), buf[:n]) == 0 {
			break
		}
		// 写入到本地文件
		if _, err := writer.Write(buf[:n]); err != nil {
			log.Logger.Error(err)
			return
		}
		if err := writer.Flush(); err != nil {
			log.Logger.Error(err)
			return
		}
	}
	DiffDeliveryOverChan <- true
}

// writeDiffData
// @Description: Client向流中写入Diff数据
// @param rw
func writeDiffData(rw *bufio.ReadWriter, s network.Stream) {
	start := time.Now()
	diffFilePath := configs.GlobalConfig.FlConfigViper.GetString("diff_path") +
		"MinerClient_" +
		strconv.Itoa(configs.ClientID) + "/" +
		"diff_epoch_" + strconv.Itoa(ReceiveGlobalEpoch) + ".dict"
	if !utils.DirectoryOrFileExists(diffFilePath) {
		log.Logger.Error("file not exists.")
		return
	}
	file, err := os.OpenFile(diffFilePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Logger.Error(err)
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	buf := make([]byte, 1024*4)
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			if _, err := rw.Write([]byte(StreamFileDeliveryFlag)); err != nil {
				log.Logger.Error(err)
				return
			}
			if err := rw.Flush(); err != nil {
				log.Logger.Error(err)
				return
			}
			break
		} else if err != nil {
			log.Logger.Error(err)
			return
		}
		_, err = rw.Write(buf[:n])
		if err != nil {
			log.Logger.Error(err)
			return
		}
		if err := rw.Flush(); err != nil {
			log.Logger.Error(err)
			return
		}
	}
	if err := s.CloseRead(); err != nil {
		log.Logger.Error(err)
		return
	}
	measure.MeasureTime(measure.DIFF, "client_" +strconv.Itoa(configs.ClientID), ReceiveGlobalEpoch, start)
	utils.ColorPrint("Close network stream.")
}

func writeModelData(rw *bufio.ReadWriter, clientID int) {
	start := time.Now()
	modelFilePath := configs.GlobalConfig.FlConfigViper.GetString("model_path") +
		"MinerClient_" + strconv.Itoa(clientID) + "/" +
		configs.GlobalConfig.FlConfigViper.GetString("model_name") + ".pth"
	if !utils.DirectoryOrFileExists(modelFilePath) {
		log.Logger.Error("file not exists.")
		return
	}
	file, err := os.OpenFile(modelFilePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Logger.Error(err)
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	buf := make([]byte, 1024*4)
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			if _, err := rw.Write([]byte(StreamFileDeliveryFlag)); err != nil {
				log.Logger.Error(err)
				return
			}
			if err := rw.Flush(); err != nil {
				log.Logger.Error(err)
				return
			}
			break
		} else if err != nil {
			log.Logger.Error(err)
			return
		}
		_, err = rw.Write(buf[:n])
		if err != nil {
			log.Logger.Error(err)
			return
		}
		if err := rw.Flush(); err != nil {
			log.Logger.Error(err)
			return
		}
	}
	measure.MeasureTime(measure.SENDMODEL, "client_" + strconv.Itoa(clientID),  GlobalEpoch-1, start)
}

func ReadModelData(rw *bufio.ReadWriter, s network.Stream) {
	dirRootPath := configs.GlobalConfig.FlConfigViper.GetString("model_path") +
		"MinerClient_" + strconv.Itoa(configs.ClientID) + "/"
	modelFilePath := dirRootPath + configs.GlobalConfig.FlConfigViper.GetString("model_name") + ".pth"
	if !utils.DirectoryOrFileExists(dirRootPath) {
		err := os.MkdirAll(dirRootPath, os.ModePerm)
		if err != nil {
			log.Logger.Error(err)
			return
		}
	}
	file, err := os.OpenFile(modelFilePath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Logger.Error(err)
		return
	}
	writer := bufio.NewWriter(file)
	buf := make([]byte, 1024*4)
	for {
		n, err := rw.Read(buf)
		if err != nil {
			log.Logger.Error(err)
			return
		}
		if bytes.Compare([]byte(StreamFileDeliveryFlag), buf[:n]) == 0 {
			break
		}
		// 写入到本地文件
		if _, err := writer.Write(buf[:n]); err != nil {
			log.Logger.Error(err)
			return
		}
		if err := writer.Flush(); err != nil {
			log.Logger.Error(err)
			return
		}
	}
	if err := s.CloseRead(); err != nil {
		log.Logger.Error(err)
		return
	}
	utils.ColorPrint("Close network stream.")
	ModelDeliveryOverChan <- true
}
