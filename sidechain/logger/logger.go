package logger

import "github.com/ipfs/go-log/v2"

var Logger = log.Logger("side-Blockchain")

func init()  {
	log.SetAllLoggers(log.LevelInfo)
}