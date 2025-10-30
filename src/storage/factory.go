package storage

import (
	"ZFS/config"
	"context"
	"fmt"
)

// NewStorage 根据配置创建存储实例
func NewStorage(ctx context.Context, cfg *config.Config) (Storage, error) {
	switch cfg.Storage.Type {
	case "local":
		root := cfg.Storage.LocalRoot
		if root == "" {
			root = "./storage" // 默认值
		}
		return NewLocalStorage(root)
		
	case "s3":
		s3cfg := S3StorageConfig{
			Bucket:          cfg.Storage.S3.Bucket,
			Region:          cfg.Storage.S3.Region,
			Prefix:          cfg.Storage.S3.Prefix,
			AccessKeyId:     cfg.Storage.S3.AccessKeyId,
			SecretAccessKey: cfg.Storage.S3.SecretAccessKey,
			Endpoint:        cfg.Storage.S3.Endpoint,
			ForcePathStyle:  cfg.Storage.S3.ForcePathStyle,
		}
		return NewS3Storage(ctx, s3cfg)
		
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", cfg.Storage.Type)
	}
}
