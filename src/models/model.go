package models

type Directory struct {
	DirID       string  `gorm:"column:dir_id;type:char(36);primaryKey"`
	Belong      string  `gorm:"column:belong;type:varchar(50);not null"`
	ParentDirID *string `gorm:"column:parent_dir_id;type:char(36)"`
	DirPath     string  `gorm:"column:dir_path;type:varchar(255);not null"`
	DirName     string  `gorm:"column:dir_name;type:varchar(100);not null"`
}

func (Directory) TableName() string {
	return "Directory"
}

type File struct {
	FileID   string `gorm:"column:file_id;type:char(36);primaryKey"`
	DirID    string `gorm:"column:dir_id;type:char(36);not null"`
	FileName string `gorm:"column:file_name;type:varchar(255);not null"`
	MD5      string `gorm:"column:md5;type:char(32);not null"`
	FileSize int64  `gorm:"column:file_size;type:bigint;not null"`
	Version  int    `gorm:"column:version;default:1"`
	Belong   string `gorm:"column:belong;type:varchar(100)"`
}

// TableName 指定 File 表名
func (File) TableName() string {
	return "File"
}

// Node 表示节点表 node
type Node struct {
	NodeID   int    `gorm:"column:node_id;primaryKey;autoIncrement"`
	NodeName string `gorm:"column:node_name;type:varchar(100)"`
	IP       string `gorm:"column:ip;type:varchar(20)"`
	Port     int    `gorm:"column:port;type:int"`
}

// TableName 指定 node 表名
func (Node) TableName() string {
	return "node"
}
