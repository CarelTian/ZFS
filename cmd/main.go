package cmd

import (
	"ZFS/config"
	"ZFS/etcd"
	"ZFS/logger"
	"ZFS/utils"
	"bufio"
	"context"
	"fmt"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"strings"
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
	etcdEndpoint := conf.Etcd.EtcdEndpoints
	serviceAddr := conf.Etcd.Address
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
	go StartServer(conf.Etcd.Address)
	ch := make(chan string)
	manager := NewManager("root", &nodes)
	mutex := make(chan struct{})
	go func() {
		mutex <- struct{}{}
	}()
	go func(r io.Reader, op chan string) {
		reader := bufio.NewReader(r)
		for {
			<-mutex
			fmt.Print(manager.prefix())
			input, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					os.Exit(0)
				} else {
					log.Fatal(err)
				}
			}
			input = strings.TrimSpace(input)
			if input == "exit" {
				fmt.Println("bye")
				os.Exit(0)
			}
			op <- input
		}
	}(os.Stdin, ch) // 立即传参执行

	for input := range ch {
		ret := manager.interpret(input)
		if len(ret) != 0 {
			fmt.Println(ret)
		}
		go func() {
			mutex <- struct{}{}
		}()
	}
	cleanup()
}
