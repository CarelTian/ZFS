package models

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

type DBManager struct {
	server   string
	username string
	password string
	port     int
	DB       *gorm.DB
}

func NewDBManager(username string, password string, server string,
	port int) *DBManager {
	DBM := &DBManager{
		server:   server,
		username: username,
		password: password,
		port:     port,
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/zfs?charset=utf8mb4&parseTime=True&loc=Local", username, password, server, port)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("faied to connect DB: %v", err)
	}
	DBM.DB = db
	return DBM
}

func (DBM *DBManager) Print() string {
	return "fuck"
}
