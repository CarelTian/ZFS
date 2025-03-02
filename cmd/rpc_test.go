package cmd

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"os"
	"path/filepath"
	"testing"

	pb "ZFS/grpc"
)

func TestListDirectory(t *testing.T) {
	// 设置测试用 storage 目录
	storagePath := "./storage"
	testDir := filepath.Join(storagePath, "testdir")

	// 创建 "./storage/testdir" 目录
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	// 测试结束后清理目录
	defer os.RemoveAll(storagePath)

	// 在 testdir 下创建一个文件
	testFilePath := filepath.Join(testDir, "file1.txt")
	file, err := os.Create(testFilePath)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	file.WriteString("hello world")
	file.Close()

	// 在 testdir 下创建一个子目录
	subDir := filepath.Join(testDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("创建子目录失败: %v", err)
	}
	f2 := filepath.Join(subDir, "file2.txt")
	file2, err := os.Create(f2)
	if err != nil {
		t.Fatalf("创建file2 失败")
	}
	file2.Close()

	// 实例化 fileServer
	s := &fileServer{}

	// 构造 ListDirectory 请求，传入相对路径 "testdir"
	req := &pb.ListDirectoryRequest{
		DirectoryPath: "testdir",
	}
	resp, err := s.ListDirectory(context.Background(), req)
	if err != nil {
		t.Fatalf("ListDirectory 调用失败: %v", err)
	}

	// 检查返回的条目是否包含 file1.txt 和 subdir
	var fileFound, dirFound bool
	for _, entry := range resp.Entries {
		fmt.Println(entry)
		if entry.Name == "file1.txt" && !entry.IsDirectory {
			fileFound = true
		}
		if entry.Name == "subdir" && entry.IsDirectory {
			dirFound = true
		}
	}

	if !fileFound {
		t.Error("未找到预期文件 file1.txt")
	}
	if !dirFound {
		t.Error("未找到预期目录 subdir")
	}
}

// dummyDownloadFileServer 用于模拟 gRPC 的 server stream
type dummyDownloadFileServer struct {
	chunks []*pb.FileChunk
	ctx    context.Context
}

// Send 将接收到的文件块存入切片中
func (d *dummyDownloadFileServer) Send(chunk *pb.FileChunk) error {
	d.chunks = append(d.chunks, chunk)
	return nil
}

// Context 返回 dummy stream 的上下文
func (d *dummyDownloadFileServer) Context() context.Context {
	return d.ctx
}

// SetHeader 设置 header，这里不做处理
func (d *dummyDownloadFileServer) SetHeader(md metadata.MD) error {
	return nil
}

// SendHeader 发送 header，这里不做处理
func (d *dummyDownloadFileServer) SendHeader(md metadata.MD) error {
	return nil
}

// SetTrailer 设置 trailer，这里不做处理
func (d *dummyDownloadFileServer) SetTrailer(md metadata.MD) {}

// SendMsg 发送消息，这里不做处理
func (d *dummyDownloadFileServer) SendMsg(m interface{}) error {
	return nil
}

// RecvMsg 接收消息，这里不做处理
func (d *dummyDownloadFileServer) RecvMsg(m interface{}) error {
	return nil
}
func TestDownloadFile(t *testing.T) {
	storageRoot := "./storage"
	// 确保 storage 目录存在
	if err := os.MkdirAll(storageRoot, 0755); err != nil {
		t.Fatalf("创建 storage 目录失败: %v", err)
	}
	// 测试结束后清理 storage 目录
	//defer os.RemoveAll(storageRoot)

	// 定义测试文件信息
	testFileName := "download_test.txt"
	testFilePath := filepath.Join(storageRoot, testFileName)
	content := []byte("This is a test content for DownloadFile testing.")
	if err := os.WriteFile(testFilePath, content, 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	// 实例化 fileServer，注意这里应确保 DownloadFile 方法内部会使用 storage 目录，
	// 即内部会做类似 filepath.Join(storageRoot, req.GetFilePath()) 的处理
	s := &fileServer{}

	// 构造 dummy stream 用于接收文件块
	dummyStream := &dummyDownloadFileServer{ctx: context.Background()}

	// 构造 DownloadFile 请求，传入相对路径（文件位于 storage 目录下）
	req := &pb.DownloadFileRequest{FilePath: testFileName}
	if err := s.DownloadFile(req, dummyStream); err != nil {
		t.Fatalf("DownloadFile 返回错误: %v", err)
	}

	// 将所有接收到的文件块拼接起来
	var result bytes.Buffer
	for _, chunk := range dummyStream.chunks {
		result.Write(chunk.Content)
	}

	// 比较拼接后的内容与预期内容是否一致
	if !bytes.Equal(result.Bytes(), content) {
		t.Errorf("下载的文件内容不匹配, got: %s, expected: %s", result.Bytes(), content)
	}
}
