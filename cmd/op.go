package cmd

import (
	pb "ZFS/grpc"
	"ZFS/utils"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	nodeName     string
	relativePath []string
	nodes        *sync.Map
	currentNode  string
	currentConn  *grpc.ClientConn
}

func ErrorMsg(msg string) string {
	return fmt.Sprintf("Error: %s", msg)
}

type Command func(manager *Manager, args []string) string

func NewManager(nodeName string, nodes *sync.Map) *Manager {
	return &Manager{
		nodeName:     nodeName,
		nodes:        nodes,
		relativePath: []string{},
	}
}

func (m *Manager) updateConnection() string {
	if len(m.relativePath) == 0 {
		if m.currentConn != nil {
			err := m.currentConn.Close()
			if err != nil {
				log.Fatal("关闭grpc连接出错")
			}
			m.currentConn = nil
			m.currentNode = ""
		}
		return ""
	}
	newNode := m.relativePath[0]
	if newNode != m.currentNode {
		if m.currentConn != nil {
			err := m.currentConn.Close()
			if err != nil {
				log.Fatal("关闭grpc连接出错")
			}
			m.currentConn = nil
		}
		addrVal, ok := m.nodes.Load(newNode)
		if !ok {
			return ErrorMsg("指定节点不存在")
		}
		addr := addrVal.(string)
		conn, err := GetConn(addr)
		if err != nil {
			log.Fatal("建立grpc连接出错")
		}
		m.currentConn = conn
		m.currentNode = newNode
	}
	return ""
}

func (m *Manager) prefix() string {
	ret := fmt.Sprintf("%s", m.nodeName)
	for _, v := range m.relativePath {
		ret += fmt.Sprintf("/%s", v)
	}
	ret += "> "
	return ret
}

func (m *Manager) interpret(command string) string {
	parts := strings.Split(command, " ")
	fn, ok := CommandMap[parts[0]]
	if !ok {
		return ErrorMsg("输入不合法")
	}
	ret := fn(m, parts[1:])
	return ret
}

func show(m *Manager, args []string) string {
	var sb strings.Builder
	if len(args) != 0 {
		return ErrorMsg("ls 输入不合法")
	}
	flag := true
	m.nodes.Range(func(key, value any) bool {
		if flag {
			sb.WriteString(fmt.Sprintf("%v: %v", key, value))
			flag = false
		} else {
			sb.WriteString(fmt.Sprintf("\n%v: %v", key, value))
		}
		return true
	})
	return sb.String()
}

func cd(m *Manager, args []string) string {
	if len(args) != 1 {
		return ErrorMsg("cd 输入不合法")
	}
	content := args[0]
	content = strings.TrimSpace(content)
	parts := strings.Split(content, "/")
	for _, path := range parts {
		if path == ".." {
			index := len(m.relativePath)
			if index > 0 {
				m.relativePath = m.relativePath[:index-1]
			}
		} else if path == "~" {
			m.relativePath = []string{}
		} else {
			m.relativePath = append(m.relativePath, path)
		}
	}
	updateErr := m.updateConnection()
	if updateErr != "" {
		index := len(m.relativePath)
		if index > 0 {
			m.relativePath = m.relativePath[:index-1]
		}
		return updateErr
	}
	return ""
}

func get(m *Manager, args []string) string {
	if len(args) != 1 {
		return ErrorMsg("get 输入不合法")
	}
	if len(m.relativePath) == 0 || m.currentConn == nil {
		return ErrorMsg("未指定节点或未建立 RPC 连接")
	}
	var remoteDir string
	if len(m.relativePath) > 1 {
		remoteDir = strings.Join(m.relativePath[1:], "/")
	}
	remotePath := args[0]
	if remoteDir != "" {
		remotePath = fmt.Sprintf("%s/%s", remoteDir, args[0])
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := pb.NewFileServiceClient(m.currentConn)
	req := &pb.DownloadFileRequest{
		FilePath: remotePath,
	}
	stream, err := client.DownloadFile(ctx, req)
	if err != nil {
		return ErrorMsg(fmt.Sprintf("远程调用出错：%v", err))
	}
	localDir := filepath.Join("data", m.currentNode)
	// 确保目录存在
	if err := os.MkdirAll(localDir, os.ModePerm); err != nil {
		return ErrorMsg(fmt.Sprintf("创建目录失败：%v", err))
	}
	localFilePath := filepath.Join(localDir, args[0])
	file, err := os.Create(localFilePath)
	if err != nil {
		return ErrorMsg(fmt.Sprintf("创建本地文件失败：%v", err))
	}
	defer file.Close()
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			file.Close()
			os.Remove(localFilePath)
			return ErrorMsg(fmt.Sprintf("接收文件数据失败：%v", err))
		}
		if _, err := file.Write(chunk.Content); err != nil {
			file.Close()
			os.Remove(localFilePath)
			return ErrorMsg(fmt.Sprintf("写入本地文件失败：%v", err))
		}
	}
	return fmt.Sprintf("文件下载成功")
}

func ls(m *Manager, args []string) string {
	var sb strings.Builder
	if len(args) != 0 {
		return ErrorMsg("ls 输入不合法")
	}
	index := len(m.relativePath)
	if index == 0 || m.currentConn == nil {
		return ErrorMsg("未指定节点")
	}
	//调用远程rpc
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := pb.NewFileServiceClient(m.currentConn)
	var path string
	for i := 1; i < len(m.relativePath); i++ {
		path += fmt.Sprintf("/%s", m.relativePath[i])
	}
	req := &pb.ListDirectoryRequest{
		DirectoryPath: path,
	}
	resp, err := client.ListDirectory(ctx, req)
	if err != nil {
		ErrorMsg(fmt.Sprintf("远程调用出错：%s", err))
	}
	flag := true
	for _, entry := range resp.Entries {
		e := entry
		var filetype rune
		isDir := e.IsDirectory
		if isDir {
			filetype = 'd'
		} else {
			filetype = '-'
		}
		if flag {
			flag = false
			sb.WriteString(fmt.Sprintf("%c  %v %v", filetype, e.Name, utils.FormatFileSize(e.Size)))
		} else {
			sb.WriteString(fmt.Sprintf("\n%c  %v %v", filetype, e.Name, utils.FormatFileSize(e.Size)))
		}

	}
	return sb.String()
}

var CommandMap = map[string]Command{
	"show": show,
	"cd":   cd,
	"ls":   ls,
	"get":  get,
}
