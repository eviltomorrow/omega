package file

import (
	"fmt"
	"net"
	"testing"

	"github.com/eviltomorrow/omega/internal/api/file/pb"
	"github.com/eviltomorrow/omega/internal/middleware"
	"github.com/eviltomorrow/omega/pkg/zlog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

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
	pb.RegisterFileServer(server, &Server{})

	go func() {
		if err := server.Serve(listen); err != nil {
			zlog.Fatal("GRPC Server startup failure", zap.Error(err))
		}
	}()
	return nil
}

func shutdownGRPC() error {
	if server == nil {
		return nil
	}
	server.Stop()
	return nil
}

func TestRunServer(t *testing.T) {
	defer shutdownGRPC()
	startupGRPC()
	select {}
}
