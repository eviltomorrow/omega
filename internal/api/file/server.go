package file

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/eviltomorrow/omega/internal/api/file/pb"
	"github.com/eviltomorrow/omega/pkg/file"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Server struct {
	pb.UnimplementedFileServer
}

// Write(File_WriteServer) error
// Read(*wrapperspb.StringValue, File_ReadServer) error
// GetInfo(context.Context, *wrapperspb.StringValue) (*Info, error)
// SetMode(context.Context, *Mode) (*emptypb.Empty, error)

func (s *Server) Write(ws pb.File_WriteServer) error {
	req, err := ws.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	if req.Path == "" {
		return fmt.Errorf("invalid path")
	}

	var (
		data = make(chan []byte, 128)
		sig  = make(chan error)
	)
	go func() {
		for {
			req, err := ws.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				sig <- err
				break
			}
			if len(req.Buf) != 0 {
				data <- req.Buf
			}
		}
		close(data)
	}()
	return file.Write(req.Path, 0644, data, sig)
}

func (s *Server) Read(req *wrapperspb.StringValue, rs pb.File_ReadServer) error {
	_, pipe, signal, err := file.Read(req.Value)
	if err != nil {
		return err
	}

loop:
	for {
		select {
		case buf, ok := <-pipe:
			if !ok {
				break loop
			}

			if err := rs.Send(&pb.Buffer{Buf: buf}); err != nil {
				return err
			}
		case err := <-signal:
			return err
		}
	}
	return nil
}

func (s *Server) GetInfo(ctx context.Context, req *wrapperspb.StringValue) (*pb.Info, error) {
	fi, err := os.Stat(req.Value)
	if err != nil && os.IsNotExist(err) {
		return &pb.Info{Exist: false}, nil
	}

	var md5 string
	if !fi.IsDir() {
		var err error
		md5, err = file.CalculateMD5(req.Value)
		if err != nil {
			return nil, err
		}
	}

	var info = &pb.Info{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    int32(fi.Mode()),
		ModTime: fi.ModTime().Unix(),
		IsDir:   fi.IsDir(),
		Md5:     md5,
	}
	return info, nil
}

func (s *Server) SetMode(ctx context.Context, req *pb.Mode) (*emptypb.Empty, error) {
	if err := os.Chmod(req.Path, fs.FileMode(req.Mode)); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
