package watchdog

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/eviltomorrow/omega/internal/api/hub"
	"github.com/eviltomorrow/omega/internal/api/watchdog/pb"
	"github.com/eviltomorrow/omega/pkg/file"
	"github.com/hashicorp/go-version"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Server struct {
	pb.UnimplementedWatchdogServer
}

const (
	S1 int = 1 // 主动停止
	S2 int = 2 // 被动停止
)

var (
	Reload      = make(chan struct{}, 1)
	Stop        = make(chan int, 1)
	Pid         = make(chan PS, 1)
	inFlightSem = make(chan struct{}, 1)
	BinFile     = "../bin/omega"
	ImageDir    = "../var/images"
)

type PS struct {
	Pid int
	Err error
}

func (s *Server) Notify(ctx context.Context, req *pb.Signal) (*wrapperspb.Int32Value, error) {
	select {
	case inFlightSem <- struct{}{}:
		defer func() { <-inFlightSem }()
	default:
		return nil, fmt.Errorf("watchdog service is busy")
	}

	select {
	case <-Pid:
	default:
	}

	switch req.Signal {
	case pb.Signal_QUIT:
		select {
		case Stop <- 1:
		default:
			return nil, fmt.Errorf("omega is stopped")
		}

	case pb.Signal_UP:
		select {
		case Reload <- struct{}{}:
		default:
			return nil, fmt.Errorf("omega is running")
		}

	default:
		return nil, fmt.Errorf("not implement signal[%v]", req.Signal)
	}

	ps := <-Pid
	return &wrapperspb.Int32Value{Value: int32(ps.Pid)}, ps.Err
}

func (s *Server) Pull(ctx context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	version, err := genVersion(req.Value)
	if err != nil {
		return nil, err
	}

	var (
		tag  = version.Original()
		path = filepath.Join(ImageDir, tag)
	)
	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}

	if err := hub.Pull(filepath.Join(path, "omega"), tag); err != nil {
		return nil, err
	}

	if err := os.Remove(BinFile); err != nil {
		return nil, err
	}

	from, err := os.Open(filepath.Join(path, "omega"))
	if err != nil {
		return nil, err
	}
	defer from.Close()

	to, err := os.OpenFile(BinFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer to.Close()

	buf := make([]byte, 4096)
	for {
		n, err := from.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if _, err := to.Write(buf[:n]); err != nil {
			return nil, err
		}
	}

	md5, err := file.CalculateMD5(BinFile)
	if err != nil {
		return nil, err
	}
	return &wrapperspb.StringValue{Value: md5}, nil
}

func genVersion(v string) (*version.Version, error) {
	if v == "" || !strings.HasPrefix(v, "v") || strings.Count(v, ".") != 2 {
		return nil, fmt.Errorf("verify tag failure, expect: v3.x.x, actual: %v", v)
	}
	version, err := version.NewVersion(v)
	if err != nil {
		return nil, err
	}
	return version, nil
}
