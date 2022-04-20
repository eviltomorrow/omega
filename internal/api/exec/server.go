package exec

import (
	"context"
	"fmt"
	"time"

	"github.com/eviltomorrow/omega/internal/api/exec/pb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	ScriptDir           = "../var/scripts"
	MaxExecTimeoutLimit = 60 * time.Second
)

type Server struct {
	pb.UnimplementedExecServer
}

func (s *Server) Exec(ctx context.Context, req *pb.C) (*wrapperspb.StringValue, error) {
	var timeout = time.Duration(req.Timeout) * time.Second
	if timeout > MaxExecTimeoutLimit {
		return nil, fmt.Errorf("exec")
	}

	switch req.Type {
	case pb.C_CMD:
		return nil, nil
	case pb.C_SHELL:
		return nil, nil
	case pb.C_PYTHON:
		return nil, nil
	default:
		return nil, nil
	}
}
