package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage S3存储实现
type S3Storage struct {
	client         *s3.Client
	bucket         string
	prefix         string // 对象key前缀
	region         string
}

// S3Config S3配置
type S3StorageConfig struct {
	Bucket          string
	Region          string
	Prefix          string
	AccessKeyId     string
	SecretAccessKey string
	Endpoint        string
	ForcePathStyle  bool
}

// NewS3Storage 创建S3存储实例
func NewS3Storage(ctx context.Context, cfg S3StorageConfig) (*S3Storage, error) {
	if cfg.Bucket == "" {
		return nil, errors.New("S3 bucket名称不能为空")
	}
	if cfg.Region == "" {
		return nil, errors.New("S3 region不能为空")
	}

	var awsCfg aws.Config
	var err error

	// 根据是否提供了访问密钥来选择认证方式
	if cfg.AccessKeyId != "" && cfg.SecretAccessKey != "" {
		// 使用提供的访问密钥
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyId,
				cfg.SecretAccessKey,
				"",
			)),
		)
	} else {
		// 使用默认凭证链（环境变量、~/.aws/credentials、IAM角色等）
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("加载AWS配置失败: %w", err)
	}

	// 创建S3客户端
	var options []func(*s3.Options)
	
	// 如果指定了自定义endpoint（例如MinIO）
	if cfg.Endpoint != "" {
		options = append(options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.ForcePathStyle
		})
	} else if cfg.ForcePathStyle {
		options = append(options, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, options...)

	// 确保prefix以/结尾（如果有的话）
	prefix := strings.TrimSpace(cfg.Prefix)
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
		prefix: prefix,
		region: cfg.Region,
	}, nil
}

// GetRoot 获取存储根路径（S3的bucket+prefix）
func (s3s *S3Storage) GetRoot() string {
	return fmt.Sprintf("s3://%s/%s", s3s.bucket, s3s.prefix)
}

// buildKey 构建完整的S3对象key
func (s3s *S3Storage) buildKey(path string) string {
	// 清理路径
	cleanPath := filepath.Clean(path)
	// 移除开头的斜杠（如果有）
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	// 将Windows路径分隔符转换为/
	cleanPath = strings.ReplaceAll(cleanPath, "\\", "/")
	
	if s3s.prefix == "" {
		return cleanPath
	}
	return s3s.prefix + cleanPath
}

// IsPathAllowed 检查路径是否在允许访问的范围内
func (s3s *S3Storage) IsPathAllowed(path string) (bool, error) {
	// S3存储中，我们通过prefix来限制访问范围
	// 检查路径是否包含".."，防止路径遍历攻击
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return false, nil
	}
	return true, nil
}

// ListDirectory 列出目录下的所有文件和子目录
func (s3s *S3Storage) ListDirectory(ctx context.Context, path string) ([]FileInfo, error) {
	// 检查路径权限
	allowed, err := s3s.IsPathAllowed(path)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errors.New("访问被拒绝：路径不合法")
	}

	// 构建S3前缀
	prefix := s3s.buildKey(path)
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	// 列出对象
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(s3s.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"), // 使用delimiter来模拟目录结构
	}

	result, err := s3s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("列出S3对象失败: %w", err)
	}

	var entries []FileInfo

	// 处理"目录"（CommonPrefixes）
	for _, prefix := range result.CommonPrefixes {
		if prefix.Prefix == nil {
			continue
		}
		// 提取目录名
		dirPath := strings.TrimSuffix(*prefix.Prefix, "/")
		dirName := filepath.Base(dirPath)
		
		entries = append(entries, FileInfo{
			Name:        dirName,
			IsDirectory: true,
			Size:        0,
		})
	}

	// 处理文件
	for _, obj := range result.Contents {
		if obj.Key == nil {
			continue
		}
		
		// 跳过目录本身（以/结尾的key）
		if strings.HasSuffix(*obj.Key, "/") {
			continue
		}
		
		// 提取文件名
		fileName := filepath.Base(*obj.Key)
		
		var size int64
		if obj.Size != nil {
			size = *obj.Size
		}
		
		entries = append(entries, FileInfo{
			Name:        fileName,
			IsDirectory: false,
			Size:        size,
		})
	}

	return entries, nil
}

// DownloadFile 下载文件，返回一个可读取的流
func (s3s *S3Storage) DownloadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	// 检查路径权限
	allowed, err := s3s.IsPathAllowed(path)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errors.New("访问被拒绝：路径不合法")
	}

	// 构建S3对象key
	key := s3s.buildKey(path)

	// 获取对象
	input := &s3.GetObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	}

	result, err := s3s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("下载S3对象失败: %w", err)
	}

	return result.Body, nil
}

// UploadFile 上传文件
func (s3s *S3Storage) UploadFile(ctx context.Context, path string, reader io.Reader) error {
	// 检查路径权限
	allowed, err := s3s.IsPathAllowed(path)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.New("访问被拒绝：路径不合法")
	}

	// 构建S3对象key
	key := s3s.buildKey(path)

	// 上传对象
	input := &s3.PutObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	}

	_, err = s3s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("上传S3对象失败: %w", err)
	}

	return nil
}

// DeleteFile 删除文件
func (s3s *S3Storage) DeleteFile(ctx context.Context, path string) error {
	// 检查路径权限
	allowed, err := s3s.IsPathAllowed(path)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.New("访问被拒绝：路径不合法")
	}

	// 构建S3对象key
	key := s3s.buildKey(path)

	// 删除对象
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	}

	_, err = s3s.client.DeleteObject(ctx, input)
	if err != nil {
		// 检查是否是NoSuchKey错误
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return fmt.Errorf("文件不存在: %s", path)
		}
		return fmt.Errorf("删除S3对象失败: %w", err)
	}

	return nil
}
