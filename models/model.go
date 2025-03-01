package models

// Directory 表示目录表 Directory
type Directory struct {
	DirID       string  `gorm:"column:dir_id;type:char(36);primaryKey"`     // 主键
	Belong      string  `gorm:"column:belong;type:varchar(50);not null"`    // 归属
	ParentDirID *string `gorm:"column:parent_dir_id;type:char(36)"`         // 父目录ID，可为空
	DirPath     string  `gorm:"column:dir_path;type:varchar(255);not null"` // 目录路径
	DirName     string  `gorm:"column:dir_name;type:varchar(100);not null"` // 目录名称
}

// TableName 指定 Directory 表名
func (Directory) TableName() string {
	return "Directory"
}

// File 表示文件表 File
type File struct {
	FileID   string `gorm:"column:file_id;type:char(36);primaryKey"`     // 主键
	DirID    string `gorm:"column:dir_id;type:char(36);not null"`        // 所属目录ID
	FileName string `gorm:"column:file_name;type:varchar(255);not null"` // 文件名称
	MD5      string `gorm:"column:md5;type:char(32);not null"`           // 文件MD5值
	FileSize int64  `gorm:"column:file_size;type:bigint;not null"`       // 文件大小
	Version  int    `gorm:"column:version;default:1"`                    // 版本号，默认1
	Belong   string `gorm:"column:belong;type:varchar(100)"`             // 归属，可选
}

// TableName 指定 File 表名
func (File) TableName() string {
	return "File"
}

// Node 表示节点表 node
type Node struct {
	NodeID   int    `gorm:"column:node_id;primaryKey;autoIncrement"` // 自增主键
	NodeName string `gorm:"column:node_name;type:varchar(100)"`      // 节点名称
	IP       string `gorm:"column:ip;type:varchar(20)"`              // IP地址
	Port     int    `gorm:"column:port;type:int"`                    // 端口号
}

// TableName 指定 node 表名
func (Node) TableName() string {
	return "node"
}
