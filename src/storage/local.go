package storage

import (
	"ZFS/utils"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage 本地文件系统存储实现
type LocalStorage struct {
	root string // 存储根目录
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(root string) (*LocalStorage, error) {
	// 确保根目录存在
	if err := os.MkdirAll(root, os.ModePerm); err != nil {
		return nil, err
	}
	
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	
	return &LocalStorage{
		root: absRoot,
	}, nil
}

// GetRoot 获取存储根路径
func (ls *LocalStorage) GetRoot() string {
	return ls.root
}

// IsPathAllowed 检查路径是否在允许访问的范围内
func (ls *LocalStorage) IsPathAllowed(path string) (bool, error) {
	fullPath := filepath.Join(ls.root, path)
	return utils.IsInStorage(ls.root, fullPath)
}

// ListDirectory 列出目录下的所有文件和子目录
func (ls *LocalStorage) ListDirectory(ctx context.Context, path string) ([]FileInfo, error) {
	fullPath := filepath.Join(ls.root, path)
	
	// 检查路径权限
	allowed, err := ls.IsPathAllowed(path)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errors.New("访问被拒绝：只能访问storage目录下的内容")
	}
	
	// 检查目录是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return []FileInfo{}, nil
	} else if err != nil {
		return nil, err
	}
	
	// 读取目录内容
	files, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	
	var entries []FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		
		entry := FileInfo{
			Name:        file.Name(),
			IsDirectory: file.IsDir(),
			Size:        info.Size(),
		}
		entries = append(entries, entry)
	}
	
	return entries, nil
}

// DownloadFile 下载文件，返回一个可读取的流
func (ls *LocalStorage) DownloadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(ls.root, path)
	
	// 检查路径权限
	allowed, err := ls.IsPathAllowed(path)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errors.New("访问被拒绝：只能下载storage目录下的文件")
	}
	
	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	
	return file, nil
}

// UploadFile 上传文件
func (ls *LocalStorage) UploadFile(ctx context.Context, path string, reader io.Reader) error {
	fullPath := filepath.Join(ls.root, path)
	
	// 检查路径权限
	allowed, err := ls.IsPathAllowed(path)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.New("访问被拒绝：只能上传到storage目录下")
	}
	
	// 确保父目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	
	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// 复制数据
	_, err = io.Copy(file, reader)
	return err
}

// DeleteFile 删除文件
func (ls *LocalStorage) DeleteFile(ctx context.Context, path string) error {
	fullPath := filepath.Join(ls.root, path)
	
	// 检查路径权限
	allowed, err := ls.IsPathAllowed(path)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.New("访问被拒绝：只能删除storage目录下的文件")
	}
	
	return os.Remove(fullPath)
}
