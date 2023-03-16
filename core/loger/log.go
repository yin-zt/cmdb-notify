package loger

import (
	log "github.com/cihub/seelog"
	config "github.com/yin-zt/cmdb-notify/core/conf"
	"os"
)

var (
	LoggerAcc log.LoggerInterface
	LoggerOpe log.LoggerInterface
)

func GetLoggerAcc() log.LoggerInterface {
	os.MkdirAll("/var/loger/", 0777)
	logger, err := log.LoggerFromConfigAsBytes([]byte(config.LogAccessConfigStr))
	if err != nil {
		log.Error("init loger fail")
		os.Exit(1)
	}
	LoggerAcc = logger
	return LoggerAcc
}

func GetLoggerOperate() log.LoggerInterface {
	os.MkdirAll("/var/loger/", 0777)
	logger, err := log.LoggerFromConfigAsBytes([]byte(config.LogOperateConfigStr))
	if err != nil {
		log.Error("init loger fail")
		os.Exit(1)
	}
	LoggerOpe = logger
	return LoggerOpe
}
