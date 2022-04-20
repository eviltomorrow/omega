package main

import (
	"fmt"
	"net"

	"github.com/eviltomorrow/omega/internal/api/file"
	"github.com/eviltomorrow/omega/internal/api/file/pb"
	"github.com/eviltomorrow/omega/internal/middleware"
	"github.com/eviltomorrow/omega/pkg/zlog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	startupGRPC()

}

var (
	Host = "0.0.0.0"
	Port = 8090

	server *grpc.Server
)

func startupGRPC() error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", Host, Port))
	if err != nil {
		return err
	}

	server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryServerRecoveryInterceptor,
			middleware.UnaryServerLogInterceptor,
		),
		grpc.ChainStreamInterceptor(
			middleware.StreamServerRecoveryInterceptor,
			middleware.StreamServerLogInterceptor,
		),
	)

	reflection.Register(server)
	pb.RegisterFileServer(server, &file.Server{})

	if err := server.Serve(listen); err != nil {
		zlog.Fatal("GRPC Server startup failure", zap.Error(err))
	}

	return nil
}
