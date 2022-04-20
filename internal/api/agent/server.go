package agent

import (
	"context"

	"github.com/eviltomorrow/omega/internal/api/agent/pb"
	"github.com/eviltomorrow/omega/internal/system"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Server struct {
	pb.UnimplementedAgentServer
}

// GetVersion(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error)
// GetSystem(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error)
// Exec(context.Context, *C) (*wrapperspb.StringValue, error)
// Ping(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error)

func (s *Server) GetVersion(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	return &wrapperspb.StringValue{Value: system.GetVersion()}, nil
}

func (s *Server) GetSystem(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	return &wrapperspb.StringValue{Value: system.GetInfo()}, nil
}

func (s *Server) Ping(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	return &wrapperspb.StringValue{Value: "pong"}, nil
}
