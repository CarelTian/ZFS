package cmd

import (
	"ZFS/config"
	"ZFS/logger"
	"ZFS/utils"
	"fmt"
	"log"
	"os"
)

func Start() {
	conf, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	logger.InitLogger(conf)
	logger.Log.Info("日志模块初始化成功")
	root := utils.InitACL("acl.yaml")
	if root != nil {

	}
	logger.Log.Info("ACL模块初始化成功")
	ch := make(chan string)
	go readInputV2(os.Stdin, ch)
	for input := range ch {
		fmt.Println("Received:", input)
	}
}
