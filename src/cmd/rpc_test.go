package cmd

import (
	"ZFS/utils"
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
	s := &FileServer{}

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
		e := entry
		fmt.Printf("%s %s\n", e.Name, utils.FormatFileSize(e.Size))

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

func (d *dummyDownloadFileServer) Context() context.Context {
	return d.ctx
}
func (d *dummyDownloadFileServer) SetHeader(md metadata.MD) error {
	return nil
}
func (d *dummyDownloadFileServer) SendHeader(md metadata.MD) error {
	return nil
}
func (d *dummyDownloadFileServer) SetTrailer(md metadata.MD) {}
func (d *dummyDownloadFileServer) SendMsg(m interface{}) error {
	return nil
}
func (d *dummyDownloadFileServer) RecvMsg(m interface{}) error {
	return nil
}

func TestDownloadFile(t *testing.T) {
	storageRoot := "./storage"
	if err := os.MkdirAll(storageRoot, 0755); err != nil {
		t.Fatalf("创建 storage 目录失败: %v", err)
	}
	testFileName := "download_test.txt"
	testFilePath := filepath.Join(storageRoot, testFileName)
	content := []byte("This is a test content for DownloadFile testing.")
	if err := os.WriteFile(testFilePath, content, 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	s := &FileServer{}

	dummyStream := &dummyDownloadFileServer{ctx: context.Background()}

	req := &pb.DownloadFileRequest{FilePath: testFileName}
	if err := s.DownloadFile(req, dummyStream); err != nil {
		t.Fatalf("DownloadFile 返回错误: %v", err)
	}

	var result bytes.Buffer
	for _, chunk := range dummyStream.chunks {
		result.Write(chunk.Content)
	}

	if !bytes.Equal(result.Bytes(), content) {
		t.Errorf("下载的文件内容不匹配, got: %s, expected: %s", result.Bytes(), content)
	}
}
