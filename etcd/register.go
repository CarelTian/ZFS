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

// RegisterService 注册服务到 etcd，并自动续租。
func RegisterService(ctx context.Context, etcdEndpoints, serviceAddr, nodeName string, ttl int64, dialTimeout int) (cleanup func(), err error) {
	// 创建 etcd 客户端
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdEndpoints},
		DialTimeout: time.Duration(dialTimeout) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("连接到 etcd 失败: %w", err)
	}

	// 创建租约
	leaseResp, err := cli.Grant(ctx, ttl)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("创建租约失败: %w", err)
	}

	// 注册服务，将 key 与租约绑定
	key := fmt.Sprintf("/services/zfs/%s", nodeName)
	_, err = cli.Put(ctx, key, serviceAddr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("注册服务失败: %w", err)
	}

	// 设置自动续租
	keepAliveChan, err := cli.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("设置自动续租失败: %w", err)
	}

	log.Printf("服务 %s 已注册，地址: %s", nodeName, serviceAddr)

	// 监控续租响应，发现问题后可采取进一步措施（例如重注册）
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

	// 返回一个回调函数，用于注销服务和关闭 etcd 客户端
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
