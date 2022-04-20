package local

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/eviltomorrow/omega/pkg/remote"
)

type Client struct {
	ptmx   *os.File
	signal chan error
}

func New(name string, ws *remote.WinSize) (remote.Terminal, error) {
	c := exec.Command(name)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		c.Dir = homeDir
	}

	ptmx, err := pty.Start(c)
	if err != nil {
		return nil, fmt.Errorf("start pty failure, nest error: %v", err)
	}

	if err := pty.Setsize(ptmx, &pty.Winsize{
		Rows: ws.Rows,
		Cols: ws.Cols,
		X:    0, Y: 0,
	}); err != nil {
		ptmx.Close()
		return nil, fmt.Errorf("set pty win size failure, nest error: %v", err)
	}

	return &Client{ptmx: ptmx}, nil
}

func (c *Client) ChangeWindow(ws *remote.WinSize) error {
	if c.ptmx == nil {
		return fmt.Errorf("panic: ptmx is nil")
	}
	return pty.Setsize(c.ptmx, &pty.Winsize{Rows: ws.Rows, Cols: ws.Cols})
}

func (c *Client) Close() error {
	if c.ptmx != nil {
		return c.ptmx.Close()
	}
	return nil
}

func (c *Client) Stdout() (io.Reader, error) {
	if c.ptmx == nil {
		return nil, fmt.Errorf("panic: ptmx is nil")
	}
	return c.ptmx, nil
}

func (c *Client) Stdin() (io.WriteCloser, error) {
	if c.ptmx == nil {
		return nil, fmt.Errorf("panic: ptmx is nil")
	}
	return c.ptmx, nil
}

func (c *Client) Stderr() (io.Reader, error) {
	return nil, nil
}

func (c *Client) Wait() error {
	return <-c.signal
}

func (c *Client) SetSignal(signal chan error) error {
	if c.signal != nil {
		return fmt.Errorf("signal already initialized")
	}

	c.signal = signal
	return nil
}
