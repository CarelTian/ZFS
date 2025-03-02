package utils

import (
	"ZFS/logger"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func InitACL(filename string) *ZFSNode {
	var root *ZFSNode
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 文件不存在，创建文件
		file, err := os.Create(filename)
		if err != nil {
			logger.Log.Error("创建acl文件失败", zap.Error(err))
			log.Fatal(err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				logger.Log.Error("关闭acl文件失败", zap.Error(err))
				log.Fatal(err)
			}
		}(file)
		return root
	} else {
		data, err := os.ReadFile(filename)
		if err != nil {
			logger.Log.Error("读取acl文件失败", zap.Error(err))
			log.Fatal(err)
		}
		err = yaml.Unmarshal(data, &root)
		if err != nil {
			logger.Log.Error("读取acl文件失败", zap.Error(err))
			log.Fatal(err)
		}
		return root
	}

}

func FormatFileSize(size int) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	if size < KB {
		return fmt.Sprintf("%dB", size)
	} else if size < MB {
		return fmt.Sprintf("%.2fKB", float64(size)/KB)
	} else if size < GB {
		return fmt.Sprintf("%.2fMB", float64(size)/MB)
	} else {
		return fmt.Sprintf("%.2fGB", float64(size)/GB)
	}
}

func IsInStorage(root, target string) (bool, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false, err
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return false, err
	}
	// 如果relative以".."开头，则说明target不在root下
	return !strings.HasPrefix(relative, ".."), nil
}

func Rename(name string, count int8) string {
	parts := strings.Split(name, ".")
	parts[0] = fmt.Sprintf("%s(%d)", parts[0], count)
	return strings.Join(parts, ".")
}

type ZFSNode struct {
	Name     string     `yaml:"name"`
	Children []*ZFSNode `yaml:"children,omitempty"`
}
