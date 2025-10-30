package storage

import (
	"context"
	"io"
)

// FileInfo 文件信息
type FileInfo struct {
	Name        string // 文件或目录名
	IsDirectory bool   // 是否为目录
	Size        int64  // 文件大小（字节）
}

// Storage 存储接口，定义统一的存储操作
type Storage interface {
	// ListDirectory 列出目录下的所有文件和子目录
	ListDirectory(ctx context.Context, path string) ([]FileInfo, error)

	// DownloadFile 下载文件，返回一个可读取的流
	DownloadFile(ctx context.Context, path string) (io.ReadCloser, error)

	// UploadFile 上传文件（可选，用于未来扩展）
	UploadFile(ctx context.Context, path string, reader io.Reader) error

	// DeleteFile 删除文件（可选，用于未来扩展）
	DeleteFile(ctx context.Context, path string) error

	// IsPathAllowed 检查路径是否在允许访问的范围内
	IsPathAllowed(path string) (bool, error)

	// GetRoot 获取存储根路径
	GetRoot() string
}
