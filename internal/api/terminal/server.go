package terminal

import (
	"strings"

	"context"
	"fmt"
	"io"
	"time"

	"github.com/eviltomorrow/omega/internal/api/terminal/pb"
	"github.com/eviltomorrow/omega/pkg/remote"
	"github.com/eviltomorrow/omega/pkg/remote/local"
	"github.com/eviltomorrow/omega/pkg/remote/ssh"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Server struct {
	pb.UnimplementedTerminalServer
}

/*
Create(context.Context, *Connection) (*wrapperspb.StringValue, error)
Exec(Terminal_ExecServer) error
ChangeWindow(context.Context, *WinSize) (*emptypb.Empty, error)
*/

func (s *Server) Create(ctx context.Context, req *pb.Connection) (*wrapperspb.StringValue, error) {
	sessionId, err := remote.GenerateTerminalSessionId()
	if err != nil {
		return nil, fmt.Errorf("generate terminal session-id failure, nest error: %v", err)
	}

	switch req.Mode {
	case pb.Connection_LOCAL:
		terminal, err := local.New("bash", &remote.WinSize{
			Rows: uint16(req.Ws.Rows),
			Cols: uint16(req.Ws.Cols),
		})
		if err != nil {
			return nil, fmt.Errorf("create local terminal failure, nest error: %v", err)
		}
		remote.SetTerminal(sessionId, terminal)

	case pb.Connection_SSH:
		if req.Resource == nil {
			return nil, fmt.Errorf("resource is nil")
		}
		var resource = req.Resource
		terminal, err := ssh.New(resource.Host, int(resource.Port), resource.Username, resource.Password, resource.Pk, &remote.WinSize{
			Rows: uint16(req.Ws.Rows),
			Cols: uint16(req.Ws.Cols),
		}, time.Duration(resource.Timeout)*time.Second)
		if err != nil {
			return nil, err
		}
		remote.SetTerminal(sessionId, terminal)

	default:
		return nil, fmt.Errorf("not support mode[%s]", req.Mode)
	}
	return &wrapperspb.StringValue{Value: sessionId}, nil
}

func (s *Server) Exec(ts pb.Terminal_ExecServer) error {
	var (
		terminal       remote.Terminal
		stdin          io.WriteCloser
		stdout, stderr io.Reader
		signal         = make(chan error, 3)
	)

	data, err := ts.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	if terminal == nil {
		terminal = remote.GetTerminal(data.SessionId)
	}
	if terminal == nil {
		return fmt.Errorf("not found terminal with session-id[%s]", data.SessionId)
	}
	terminal.SetSignal(signal)

	defer func() {
		remote.DelTerminal(data.SessionId)

		if terminal != nil {
			terminal.Close()
		}
	}()

	if stdin, err = terminal.Stdin(); err != nil {
		return err
	}
	if stdout, err = terminal.Stdout(); err != nil {
		return err
	}
	if stderr, err = terminal.Stderr(); err != nil {
		return err
	}

	go read(stdout, ts, signal)
	go read(stderr, ts, signal)
	go write(stdin, ts, signal)

	return terminal.Wait()
}

func (s *Server) ChangeWindow(ctx context.Context, req *pb.WinSize) (*emptypb.Empty, error) {
	var (
		sessionId = req.SessionId
		terminal  = remote.GetTerminal(sessionId)
	)
	if terminal == nil {
		return nil, fmt.Errorf("not found terminal with session-id[%s]", sessionId)
	}
	if err := terminal.ChangeWindow(&remote.WinSize{
		Rows: uint16(req.Rows),
		Cols: uint16(req.Cols),
	}); err != nil {
		return nil, fmt.Errorf("change window size failure, nest error: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func write(writer io.WriteCloser, ts pb.Terminal_ExecServer, signal chan error) {
	for {
		data, err := ts.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		if _, err := writer.Write(data.Buf); err != nil {
			signal <- err
			break
		}
	}
}

func read(reader io.Reader, ts pb.Terminal_ExecServer, signal chan error) {
	if reader != nil {
		var buf [4096]byte
		for {
			n, err := reader.Read(buf[0:])
			if err == io.EOF {
				signal <- io.EOF
				break
			}
			if err != nil {
				if strings.Contains(err.Error(), "input/output error") {
					signal <- io.EOF
				} else {
					signal <- err
				}
				break
			}
			if err := ts.Send(&pb.Data{Buf: buf[:n]}); err != nil {
				break
			}
		}
	}
}
