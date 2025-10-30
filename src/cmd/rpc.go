package cmd

import (
	pb "ZFS/grpc"
	"ZFS/storage"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net"
)

type FileServer struct {
	pb.UnimplementedFileServiceServer
	storage storage.Storage
}

type FileService struct{}

func StartServer(addr string, stor storage.Storage) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()

	pb.RegisterFileServiceServer(grpcServer, &FileServer{storage: stor})
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
	
	// 使用storage层列出目录
	files, err := s.storage.ListDirectory(ctx, dirPath)
	if err != nil {
		return nil, err
	}
	
	// 转换为protobuf格式
	var entries []*pb.FileEntry
	for _, file := range files {
		entry := &pb.FileEntry{
			Name:        file.Name,
			IsDirectory: file.IsDirectory,
			Size:        file.Size,
		}
		entries = append(entries, entry)
	}
	
	return &pb.ListDirectoryResponse{Entries: entries}, nil
}

func (s *FileServer) DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error {
	filePath := req.GetFilePath()
	
	// 使用storage层下载文件
	reader, err := s.storage.DownloadFile(stream.Context(), filePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 流式传输文件内容
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
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
