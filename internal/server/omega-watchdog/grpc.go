package server

import (
	"fmt"
	"log"
	"net"

	"github.com/eviltomorrow/omega/internal/api/watchdog"
	pb_watchdog "github.com/eviltomorrow/omega/internal/api/watchdog/pb"
	"github.com/eviltomorrow/omega/internal/middleware"
	"github.com/eviltomorrow/omega/pkg/grpclb"
	"github.com/eviltomorrow/omega/pkg/tools"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	BinFile        = "../bin/omega"
	ImageDir       = "../var/images"
	Key            = "grpclb/service/omega-watchdog/unknown"
	InnerIP        = ""
	OuterIP        = ""
	Port           = 28500
	Endpoints      = []string{}
	RevokeEtcdConn func() error

	server *grpc.Server
)

func StartupGRPC() error {
	if InnerIP == "" {
		var err error
		InnerIP, err = tools.GetLocalIP2()
		if err != nil {
			return fmt.Errorf("get local ip failure, nest error: %v", err)
		}
		if OuterIP == "" {
			OuterIP = InnerIP
		}
	}

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", InnerIP, Port))
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

	pb_watchdog.RegisterWatchdogServer(server, &watchdog.Server{})

	close, err := grpclb.Register(Key, InnerIP, OuterIP, Port, Endpoints, 10)
	if err != nil {
		return fmt.Errorf("register service to etcd failure, nest error: %v", err)
	}
	RevokeEtcdConn = func() error {
		close()
		return nil
	}

	go func() {
		if err := server.Serve(listen); err != nil {
			log.Fatalf("[F] GRPC Server startup failure, nest error: %v", err)
		}
	}()
	return nil
}

func ShutdownGRPC() error {
	if server == nil {
		return nil
	}
	server.Stop()
	return nil
}
