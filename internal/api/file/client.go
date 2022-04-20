package file

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/eviltomorrow/omega/internal/api/file/pb"
	"github.com/eviltomorrow/omega/pkg/bar"
	"github.com/eviltomorrow/omega/pkg/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func NewClient(target string) (pb.FileClient, func(), error) {
	conn, err := grpc.DialContext(
		context.Background(),
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("dial %s failure, nest error: %v", target, err)
	}

	return pb.NewFileClient(conn), func() { conn.Close() }, nil
}

func Upload(target string, local string, remote string) error {
	localFi, err := os.Stat(local)
	if err != nil {
		return err
	}
	if localFi.IsDir() {
		return fmt.Errorf("local: %s is a folder", local)
	}
	localMD5, err := file.CalculateMD5(local)
	if err != nil {
		return err
	}

	stub, destroy, err := NewClient(target)
	if err != nil {
		return err
	}
	defer destroy()

	remoteFi, err := stub.GetInfo(context.Background(), &wrapperspb.StringValue{Value: remote})
	if err != nil {
		return err
	}
	if remoteFi.IsDir {
		remote = filepath.Join(remote, filepath.Base(local))
		remoteFi, err = stub.GetInfo(context.Background(), &wrapperspb.StringValue{Value: remote})
		if err != nil {
			return err
		}
		if remoteFi.IsDir {
			return fmt.Errorf("one folder with the same name exists, name: %v", remote)
		}
	}
	if remoteFi.Md5 == localMD5 {
		return nil
	}

	localF, err := os.OpenFile(local, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer localF.Close()

	writer, err := stub.Write(context.Background())
	if err != nil {
		return err
	}

	if err := writer.Send(&pb.Buffer{Path: remote}); err != nil {
		return err
	}

	bar, counter := bar.NewProgressbar(int(remoteFi.Size), "Upload", remote)
	defer bar.Close()
	defer close(counter)

	var buf [1024 * 8]byte
	for {
		n, err := localF.Read(buf[0:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := writer.Send(&pb.Buffer{Buf: buf[:n]}); err != nil {
			return err
		}
		counter <- n
	}
	if _, err := writer.CloseAndRecv(); err != nil && err != io.EOF {
		return err
	}

	if _, err := stub.SetMode(context.Background(), &pb.Mode{Path: remote, Mode: int32(localFi.Mode())}); err != nil {
		return err
	}

	return nil
}

func Download(target string, remote string, local string) error {
	stub, destroy, err := NewClient(target)
	if err != nil {
		return err
	}
	defer destroy()

	remoteFi, err := stub.GetInfo(context.Background(), &wrapperspb.StringValue{Value: remote})
	if err != nil {
		return err
	}
	if remoteFi.IsDir {
		return fmt.Errorf("remote path is a folder")
	}

	localFi, err := os.Stat(local)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		if localFi.IsDir() {
			local = filepath.Join(local, filepath.Base(remoteFi.Name))
			localFi, err = os.Stat(local)
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			if err == nil {
				if localFi.IsDir() {
					return fmt.Errorf("one folder with the same name exists, name: %v", local)
				}
			}
		}
		localMD5, err := file.CalculateMD5(local)
		if err != nil {
			return err
		}

		if localMD5 == remoteFi.Md5 {
			return nil
		}
	}

	reader, err := stub.Read(context.Background(), &wrapperspb.StringValue{Value: remote})
	if err != nil {
		return err
	}

	var (
		pipe   = make(chan []byte, 128)
		signal = make(chan error, 1)
	)

	go func() {
		bar, counter := bar.NewProgressbar(int(remoteFi.Size), "Download", remote)
		defer bar.Close()
		defer close(counter)

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
			counter <- len(data.Buf)
		}
		close(pipe)
	}()

	if err := file.Write(local, fs.FileMode(remoteFi.Mode), pipe, signal); err != nil {
		return err
	}

	return nil
}
