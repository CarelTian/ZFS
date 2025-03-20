package etcd_test

import (
	"ZFS/etcd"
	"bytes"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"testing"
	"time"
)

func startEmbeddedEtcd(t *testing.T, timeout time.Duration) *embed.Etcd {
	t.Helper()
	cfg := embed.NewConfig()

	dir, err := ioutil.TempDir("", "etcd-"+t.Name())
	if err != nil {
		t.Fatalf("创建临时目录失败: %v, 路径: %s", err, dir)
	}
	t.Logf("创建临时目录: %s", dir)
	cfg.Dir = dir

	lp, err := url.Parse("http://localhost:2379")
	if err != nil {
		t.Fatalf("解析地址失败: %v", err)
	}
	cfg.ListenClientUrls = []url.URL{*lp}
	cfg.AdvertiseClientUrls = cfg.ListenClientUrls

	ep, err := url.Parse("http://localhost:2380")
	if err != nil {
		t.Fatalf("解析地址失败: %v", err)
	}
	cfg.ListenPeerUrls = []url.URL{*ep}
	cfg.AdvertisePeerUrls = cfg.ListenPeerUrls

	t.Logf("开始启动etcd，配置: %+v", cfg)

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		t.Fatalf("启动嵌入式 etcd 失败: %v", err)
	}

	select {
	case <-e.Server.ReadyNotify():
		t.Log("嵌入式 etcd 已启动")
	case <-time.After(timeout):
		e.Server.Stop()
		t.Fatal("嵌入式 etcd 启动超时")
	}

	t.Cleanup(func() {
		t.Log("关闭 etcd 并清理资源")
		e.Close()
		os.RemoveAll(cfg.Dir)
	})

	return e
}

func TestRegisterService(t *testing.T) {
	e := startEmbeddedEtcd(t, 10*time.Second)

	ep := e.Clients[0].Addr().String()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nodeName := "Node1"
	serviceAddr := "http://127.0.0.1:8000"
	ttl := int64(5)
	dialTimeout := 5

	cleanup, err := etcd.RegisterService(ctx, ep, serviceAddr, nodeName, ttl, dialTimeout)
	if err != nil {
		t.Fatalf("RegisterService 调用失败: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{ep},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("创建 etcd 客户端失败: %v", err)
	}
	defer cli.Close()

	key := fmt.Sprintf("/services/zfs/%s", nodeName)
	resp, err := cli.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("获取 key 失败: %v", err)
	}
	if len(resp.Kvs) == 0 {
		t.Fatalf("预期 key %s 存在", key)
	}
	if string(resp.Kvs[0].Value) != serviceAddr {
		t.Fatalf("预期值 %s，实际获得 %s", serviceAddr, string(resp.Kvs[0].Value))
	}

	cleanup()
	time.Sleep(500 * time.Millisecond)
	resp, err = cli.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("cleanup 后获取 key 失败: %v", err)
	}
	if len(resp.Kvs) != 0 {
		t.Fatalf("预期 key %s 被移除", key)
	}
}

func TestDiscoverService(t *testing.T) {
	e := startEmbeddedEtcd(t, 10*time.Second)

	ep := e.Clients[0].Addr().String()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{ep},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("创建 etcd 客户端失败: %v", err)
	}
	defer cli.Close()

	var buf bytes.Buffer
	origOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(origOutput)

	key1 := "/services/zfs/test-node1"
	val1 := "http://127.0.0.1:9000"
	_, err = cli.Put(context.Background(), key1, val1)
	if err != nil {
		t.Fatalf("预写入 key 失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan error, 1)
	go func() {
		err := etcd.DiscoverService(ctx, ep)
		doneCh <- err
	}()

	time.Sleep(500 * time.Millisecond)
	logs := buf.String()
	if !contains(logs, "Initial service found") {
		t.Fatalf("未检测到初始同步日志，实际日志: %s", logs)
	}

	buf.Reset()

	key2 := "/services/zfs/test-node2"
	val2 := "http://127.0.0.1:9100"
	_, err = cli.Put(context.Background(), key2, val2)
	if err != nil {
		t.Fatalf("写入新 key 失败: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
	logs = buf.String()
	if !contains(logs, "Service added/updated") {
		t.Fatalf("未检测到服务添加日志，实际日志: %s", logs)
	}

	// 清空日志缓冲
	buf.Reset()

	// 删除新写入的 key 以触发删除事件
	_, err = cli.Delete(context.Background(), key2)
	if err != nil {
		t.Fatalf("删除 key 失败: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
	logs = buf.String()
	if !contains(logs, "Service removed") {
		t.Fatalf("未检测到服务删除日志，实际日志: %s", logs)
	}

	// 停止 DiscoverService
	cancel()
	select {
	case err := <-doneCh:
		if err != context.Canceled && err != nil {
			t.Fatalf("DiscoverService 非正常退出: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("DiscoverService 停止超时")
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
