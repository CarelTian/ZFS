package cmd

import (
	pb "ZFS/grpc"
	"ZFS/utils"
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net"
	"os"
	"path/filepath"
)

type FileServer struct {
	pb.UnimplementedFileServiceServer
}

type FileService struct{}

const root = "./storage"

func StartServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()

	pb.RegisterFileServiceServer(grpcServer, &FileServer{})
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}

func GetConn(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// 实现 ListDirectory 方法
func (s *FileServer) ListDirectory(ctx context.Context, req *pb.ListDirectoryRequest) (*pb.ListDirectoryResponse, error) {

	dirPath := req.GetDirectoryPath()
	fullPath := filepath.Join(root, dirPath)
	var entries []*pb.FileEntry

	inStorage, err := utils.IsInStorage(root, fullPath)
	if err != nil {
		return nil, err
	}
	if !inStorage {
		return nil, errors.New("访问被拒绝：只能访问storage目录下的内容")
	}
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return &pb.ListDirectoryResponse{Entries: entries}, nil
	} else if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		entry := &pb.FileEntry{
			Name:        file.Name(),
			IsDirectory: file.IsDir(),
			Size:        info.Size(),
		}
		entries = append(entries, entry)
	}
	return &pb.ListDirectoryResponse{Entries: entries}, nil
}

func (s *FileServer) DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error {
	filePath := filepath.Join(root, req.GetFilePath())
	inStorage, err := utils.IsInStorage(root, filePath)
	if err != nil {
		return err
	}
	if !inStorage {
		return errors.New("访问被拒绝：只能下载storage目录下的文件")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		chunk := &pb.FileChunk{
			Content: buf[:n],
		}
		if err := stream.Send(chunk); err != nil {
			return err
		}
	}
	return nil

}
