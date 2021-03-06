package file

import (
	"context"
	"io"
	"os"
	"time"

	"golang.org/x/time/rate"
)

func Write(path string, mode os.FileMode, data chan []byte, signal chan error) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	for {
		select {
		case buf, ok := <-data:
			if !ok {
				return nil
			}
			if _, err := f.Write(buf); err != nil {
				return err
			}
		case err := <-signal:
			return err
		}
	}
}

func Read(path string) (os.FileInfo, chan []byte, chan error, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, nil, nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, nil, err
	}

	var (
		p      = make(chan []byte, 32)
		signal = make(chan error, 1)
	)

	go func() {
		defer f.Close()

		var limiter = rate.NewLimiter(rate.Every(time.Second/1000), 1)
		for {
			var buf [1024 * 8]byte
			n, err := f.Read(buf[0:])
			if err == io.EOF {
				break
			}
			if err != nil {
				signal <- err
				break
			}
			limiter.WaitN(context.Background(), 1)

			p <- buf[:n]
		}
		close(p)
	}()
	return fi, p, signal, nil
}
