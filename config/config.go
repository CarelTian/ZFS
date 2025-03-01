package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Node    NodeConfig                 `yaml:"node"`
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
