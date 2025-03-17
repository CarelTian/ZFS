package etcd

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const servicePrefix = ""

// 通过 Watch 机制监控 etcd 中的服务变化
func DiscoverService(ctx context.Context, etcdEndpoints string, nodes *sync.Map) error {
	// 创建 etcd 客户端
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdEndpoints},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("连接到 etcd 失败: %w", err)
	}
	defer cli.Close()

	// 初始同步：读取当前所有注册的服务
	resp, err := cli.Get(ctx, servicePrefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("获取初始服务列表失败: %w", err)
	}

	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		value := string(kv.Value)
		nodes.Store(key, value)
		//log.Printf("发现初始服务: %s -> %s", kv.Key, kv.Value)
	}

	// 使用 Watch 监控服务目录的变化
	watchChan := cli.Watch(ctx, servicePrefix, clientv3.WithPrefix())
	//log.Println("开始监视服务注册表...")
	for {
		select {
		case <-ctx.Done():
			log.Println("服务发现已停止")
			return ctx.Err()
		case watchResp, ok := <-watchChan:
			if !ok {
				log.Println("Watch 频道已关闭")
				return nil
			}
			if watchResp.Err() != nil {
				log.Printf("Watch 错误: %v", watchResp.Err())
				continue
			}
			for _, ev := range watchResp.Events {
				key := string(ev.Kv.Key)
				switch ev.Type {
				case clientv3.EventTypePut:
					value := string(ev.Kv.Value)
					nodes.Store(key, value)
					log.Printf("服务添加/更新: %s -> %s", ev.Kv.Key, ev.Kv.Value)
				case clientv3.EventTypeDelete:
					nodes.Delete(key)
					log.Printf("服务移除: %s", ev.Kv.Key)
				}
			}
		}
	}
}
