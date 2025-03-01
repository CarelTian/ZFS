package utils

import (
	"ZFS/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"log"
	"os"
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

type ZFSNode struct {
	Name     string     `yaml:"name"`
	Children []*ZFSNode `yaml:"children,omitempty"`
}
