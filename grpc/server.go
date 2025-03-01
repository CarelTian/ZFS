package grpc

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type FileService struct{}

func StartServer(conf struct{ Port int }) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	// 注册服务
	// pb.RegisterFileServiceServer(grpcServer, &FileService{})
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
