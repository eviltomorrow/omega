package terminal

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/eviltomorrow/omega/internal/api/terminal/pb"
	"github.com/eviltomorrow/omega/internal/conf"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Resource struct {
	Username       string        `json:"username"`
	Password       string        `json:"password"`
	Host           string        `json:"host"`
	Port           int           `json:"port"`
	PrivateKeyPath string        `json:"pk_file"`
	Timeout        conf.Duration `json:"timeout"`
}

func NewClient(target string) (pb.TerminalClient, func(), error) {
	conn, err := grpc.DialContext(
		context.Background(),
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("dial %s failure, nest error: %v", target, err)
	}

	return pb.NewTerminalClient(conn), func() { conn.Close() }, nil
}

func NewSSH(name string, target string, timeout time.Duration, resource *Resource) error {
	var pk []byte
	if resource.PrivateKeyPath != "" {
		buf, err := ioutil.ReadFile(resource.PrivateKeyPath)
		if err != nil {
			return err
		}
		pk = buf
	}
	return newTerminal(name, target, timeout, &pb.Connection{
		Mode: pb.Connection_SSH,
		Resource: &pb.Resource{
			Username: resource.Username,
			Password: resource.Password,
			Host:     resource.Host,
			Pk:       pk,
			Port:     int32(resource.Port),
			Timeout:  int32(resource.Timeout.Duration),
		},
	})
}

func NewLocal(name, target string, timeout time.Duration) error {
	return newTerminal(name, target, timeout, &pb.Connection{Mode: pb.Connection_LOCAL})
}

func newTerminal(name, target string, timeout time.Duration, connection *pb.Connection) error {
	stub, close, err := NewClient(target)
	if err != nil {
		return err
	}
	defer close()

	size, err := pty.GetsizeFull(os.Stdin)
	if err != nil {
		return err
	}

	connection.Ws = &pb.WinSize{
		Rows: int32(size.Rows),
		Cols: int32(size.Cols),
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := stub.Create(ctx, connection)
	if err != nil {
		return err
	}
	var sessionId = resp.Value

	pipe, err := stub.Exec(context.Background())
	if err != nil {
		return err
	}

	if err := pipe.Send(&pb.Data{SessionId: sessionId}); err != nil {
		return err
	}

	c := exec.Command(name)
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}

	defer func() { _ = ptmx.Close() }()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			size, err := pty.GetsizeFull(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
			if err := pty.Setsize(ptmx, size); err != nil {
				log.Fatal(err)
			}
			stub.ChangeWindow(context.Background(), &pb.WinSize{SessionId: sessionId, Rows: int32(size.Rows), Cols: int32(size.Cols)})
		}
	}()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() {
		var buf [4096]byte

		defer func() {
			pipe.CloseSend()
		}()
		for {
			n, err := os.Stdin.Read(buf[0:])
			if err != nil {
				log.Printf("panic: read stdin failure, nest error: %v\r\n", err)
				return
			}
			err = pipe.Send(&pb.Data{SessionId: sessionId, Buf: buf[:n]})
			if err != nil {
				log.Printf("panic: send data to pipe failure, nest error: %v\r\n", err)
				return
			}
		}
	}()

	defer func() {
		pipe.CloseSend()
	}()
	for {
		data, err := pipe.Recv()
		if err != nil {
			if s := status.Convert(err); s != nil && s.Message() == "EOF" {
				break
			}
			log.Printf("panic: recv data from pipe failure, nest error: %v\r\n", err)
			return nil
		}
		_, err = os.Stdout.Write(data.Buf)
		if err != nil {
			log.Printf("panic: write data to stdout failure, nest error: %v\r\n", err)
			return nil
		}
	}
	return nil
}
