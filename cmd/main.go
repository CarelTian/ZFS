package cmd

import (
	"ZFS/config"
	"ZFS/etcd"
	"ZFS/logger"
	"ZFS/utils"
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
	"sync"
)

var nodes sync.Map

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
	etcdEndpoint := conf.Etcd.EtcdEndpoints // 例如 "localhost:2379"
	serviceAddr := conf.Etcd.Address        // 例如 "http://localhost:8080"
	ttl := conf.Etcd.TTL
	dialTimeout := conf.Etcd.DialTimeout
	nodeName := conf.Node.Name

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cleanup, err := etcd.RegisterService(ctx, etcdEndpoint, serviceAddr, nodeName, int64(ttl), dialTimeout)
	if err != nil {
		logger.Log.Error("服务注册失败", zap.Error(err))
		log.Fatalf("服务注册失败: %v", err)
	}
	go func() {
		if err := etcd.DiscoverService(ctx, etcdEndpoint, &nodes); err != nil {
			logger.Log.Error("服务发现异常", zap.Error(err))
		}
	}()
	ch := make(chan string)
	go readInputV2(os.Stdin, ch)
	for input := range ch {
		if input == "ls" {
			nodes.Range(func(key, value any) bool {
				fmt.Println(key, value)
				return true
			})
		} else {
			fmt.Println("Received:", input)
		}

	}

	cleanup()
}
