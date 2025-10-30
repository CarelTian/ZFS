package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Node    NodeConfig                 `yaml:"node"`
	Storage StorageConfig              `yaml:"storage"`
	Etcd    EtcdConfig                 `yaml:"etcd"`
	MySQL   MysqlConfig                `yaml:"mysql"`
	Kafka   struct{ Brokers []string } `yaml:"kafka"`
	GRPC    struct{ Port int }         `yaml:"grpc"`
	Gateway struct{ Port int }         `yaml:"gateway"`
	Log     LogConfig                  `yaml:"log"`
}

type LogConfig struct {
	Enable           bool     `yaml:"enable"`
	Level            string   `yaml:"level"`
	Encoding         string   `yaml:"encoding"`
	OutputPaths      []string `yaml:"outputPaths"`
	ErrorOutputPaths []string `yaml:"errorOutputPaths"`
}

type MysqlConfig struct {
	Server   string `yaml:"server"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
}

type EtcdConfig struct {
	EtcdEndpoints string `yaml:"etcdEndpoints"` // etcd 服务端点
	Address       string `yaml:"address"`       // etcd 对外提供的地址
	TTL           int    `yaml:"ttl"`           // 键的生存时间（秒）
	DialTimeout   int    `yaml:"dialTimeout"`   // 连接超时（秒）
}

type NodeConfig struct {
	Name    string `yaml:"name"` // 节点名字
	Storage string `yaml:"storage"`
}

type StorageConfig struct {
	Type      string      `yaml:"type"`      // 存储类型：local 或 s3
	LocalRoot string      `yaml:"localRoot"` // 本地存储根目录
	DataRoot  string      `yaml:"dataRoot"`  // 下载文件保存目录
	S3        S3Config    `yaml:"s3"`        // S3配置
}

type S3Config struct {
	Bucket           string `yaml:"bucket"`           // S3存储桶名称
	Region           string `yaml:"region"`           // AWS区域
	Prefix           string `yaml:"prefix"`           // 对象key前缀
	AccessKeyId      string `yaml:"accessKeyId"`      // 访问密钥ID
	SecretAccessKey  string `yaml:"secretAccessKey"`  // 访问密钥
	Endpoint         string `yaml:"endpoint"`         // 自定义endpoint（用于MinIO等）
	ForcePathStyle   bool   `yaml:"forcePathStyle"`   // 是否使用路径风格访问
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
