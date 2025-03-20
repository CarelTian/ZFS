package etcd

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

//etcd --listen-client-urls http://localhost:2379 \
//--advertise-client-urls http://localhost:2379 \
//--listen-peer-urls http://localhost:2380 \
//--initial-advertise-peer-urls http://localhost:2380

func RegisterService(ctx context.Context, etcdEndpoints, serviceAddr, nodeName string, ttl int64, dialTimeout int) (cleanup func(), err error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdEndpoints},
		DialTimeout: time.Duration(dialTimeout) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("连接到 etcd 失败: %w", err)
	}

	leaseResp, err := cli.Grant(ctx, ttl)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("创建租约失败: %w", err)
	}

	key := fmt.Sprintf("%s", nodeName)
	_, err = cli.Put(ctx, key, serviceAddr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("注册服务失败: %w", err)
	}

	keepAliveChan, err := cli.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("设置自动续租失败: %w", err)
	}

	log.Printf("服务 %s 已注册，地址: %s", nodeName, serviceAddr)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("上下文取消，停止节点 %s 的自动续租", nodeName)
				return
			case _, ok := <-keepAliveChan:
				if !ok {
					log.Printf("节点 %s 的 KeepAlive 频道已关闭", nodeName)
					// 这里可以添加重注册逻辑或告警处理
					return
				}
				//log.Printf("收到节点 %s 的续租响应，TTL: %d", nodeName, ka.TTL)
			}
		}
	}()

	cleanup = func() {
		_, err := cli.Delete(context.Background(), key)
		if err != nil {
			log.Printf("删除键 %s 时出错: %v", key, err)
		}
		cli.Close()
		log.Printf("服务 %s 已注销", nodeName)
	}

	return cleanup, nil
}
