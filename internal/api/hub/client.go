package hub

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eviltomorrow/omega/internal/api/hub/pb"
	"github.com/eviltomorrow/omega/pkg/bar"
	"github.com/eviltomorrow/omega/pkg/exec"
	"github.com/eviltomorrow/omega/pkg/file"
	"github.com/eviltomorrow/omega/pkg/self"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	DefaultDialTimeout = 5 * time.Second
)

func newHubConn() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultDialTimeout)
	defer cancel()

	target := fmt.Sprintf("etcd:///%s/omega-hub", self.EtcdKeyPrefix)
	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("dial %s failure, nest error: %v", target, err)
	}
	return conn, nil
}

func NewClient() (pb.HubClient, func(), error) {
	conn, err := newHubConn()
	if err != nil {
		return nil, nil, err
	}

	return pb.NewHubClient(conn), func() { conn.Close() }, nil
}

func Push(local string, releaseNote string) (string, error) {
	tag, created, err := verifyOmega(local)
	if err != nil {
		return "", err
	}

	stub, destroy, err := NewClient()
	if err != nil {
		return "", err
	}
	defer destroy()

	localF, err := os.OpenFile(local, os.O_RDONLY, 0644)
	if err != nil {
		return "", err
	}
	defer localF.Close()

	localFI, err := localF.Stat()
	if err != nil {
		return "", err
	}

	writer, err := stub.Push(context.Background())
	if err != nil {
		return "", err
	}

	if err := writer.Send(&pb.Image{ReleaseNotes: releaseNote, Tag: tag, CreateTime: created}); err != nil {
		return "", err
	}

	p, counter := bar.NewProgressbar(int(localFI.Size()), "Push", fmt.Sprintf("image [omega-%s]", tag))
	defer p.Close()
	defer close(counter)

	var (
		buf     [1024 * 8]byte
		limiter = rate.NewLimiter(rate.Every(time.Second/1000), 1)
	)

	for {
		n, err := localF.Read(buf[0:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		limiter.WaitN(context.Background(), 1)

		if err := writer.Send(&pb.Image{Buf: buf[:n]}); err != nil {
			return "", err
		}
		counter <- n
	}

	if err := writer.CloseSend(); err != nil {
		return "", err
	}

	resp, err := writer.CloseAndRecv()
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func Pull(local string, tag string) error {
	stub, destroy, err := NewClient()
	if err != nil {
		return err
	}
	defer destroy()

	localFi, err := os.Stat(local)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil && localFi.IsDir() {
		return fmt.Errorf("panic: local file is a dir")
	}

	reader, err := stub.Pull(context.Background(), &wrapperspb.StringValue{Value: tag})
	if err != nil {
		return err
	}

	var (
		pipe   = make(chan []byte, 128)
		signal = make(chan error, 1)
	)

	go func() {
		for {
			data, err := reader.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				signal <- err
				return
			}

			pipe <- data.Buf

		}
		close(pipe)
	}()

	if err := file.Write(local, 0755, pipe, signal); err != nil {
		os.Remove(local)
		return err
	}
	return nil
}

func verifyOmega(local string) (string, string, error) {
	localFi, err := os.Stat(local)
	if err != nil {
		return "", "", err
	}
	if localFi.IsDir() {
		return "", "", fmt.Errorf("panic: local file is a dir")
	}

	var (
		base = filepath.Dir(local)
		name = filepath.Base(local)
	)

	stdout, _, err := exec.RunCmd(fmt.Sprintf("cd %s; ./%s version", base, name), 3*time.Second)
	if err != nil {
		return "", "", err
	}

	var (
		reader      = strings.NewReader(stdout)
		scanner     = bufio.NewScanner(reader)
		mainVersion string
		tag         string
		buildTime   string
	)

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var text = scanner.Text()
		if strings.Contains(text, "Omega Version (Current)") {
			if strings.Contains(text, ":") {
				var attr = strings.Split(text, ":")
				if len(attr) != 2 {
					return "", "", fmt.Errorf("unknown main version format")
				}
				mainVersion = strings.TrimSpace(attr[1])
			}
		}
		if strings.Contains(text, "Git Tag") {
			if strings.Contains(text, ":") {
				var attr = strings.Split(text, ":")
				if len(attr) != 2 {
					return "", "", fmt.Errorf("unknown git tag format")
				}
				tag = strings.TrimSpace(attr[1])
			}
		}
		if strings.Contains(text, "Build Time:") {
			var attr = strings.Split(text, "Build Time:")
			if len(attr) != 2 {
				return "", "", fmt.Errorf("unknown build time")
			}
			buildTime = strings.TrimSpace(attr[1])
		}
	}

	if !strings.HasPrefix(tag, "v") {
		tag = mainVersion
	}

	version, err := genVersion(tag)
	if err != nil {
		return "", "", err
	}

	return version.Original(), buildTime, nil
}
