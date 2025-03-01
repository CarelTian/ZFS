package models

import (
	"ZFS/config"
	"log"
	"testing"
)

func TestConnection(t *testing.T) {
	conf, err := config.LoadConfig("../config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	_ = NewDBManager(conf.MySQL.Username, conf.MySQL.Password, conf.MySQL.Server, conf.MySQL.Port)
}
