package ssh

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/eviltomorrow/omega/pkg/remote"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	conn    *ssh.Client
	session *ssh.Session

	stdin          io.WriteCloser
	stdout, stderr io.Reader
}

func New(host string, port int, username, password string, pk []byte, ws *remote.WinSize, timeout time.Duration) (remote.Terminal, error) {
	var authMethods = make([]ssh.AuthMethod, 0, 4)
	if len(pk) != 0 {
		signer, err := ssh.ParsePrivateKey(pk)
		if err != nil {
			return nil, fmt.Errorf("parse private key failure, nest error: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if password != "" {
		authMethods = append(authMethods, ssh.KeyboardInteractive(setKeyboard(password)))
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("panic: no valid auth method is included, nest error: password or pk may not be exist")
	}

	config := ssh.ClientConfig{
		User: username,
		Auth: authMethods,
		Config: ssh.Config{
			Ciphers: []string{
				"aes128-ctr",
				"aes192-ctr",
				"aes256-ctr",
				"aes128-gcm@openssh.com",
				"arcfour256",
				"arcfour128",
				"aes128-cbc",
			},
		},
		Timeout: timeout,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)), &config)
	if err != nil {
		return nil, fmt.Errorf("dial [%s@%s:%d] failure, nest error: %v", username, host, port, err)
	}

	session, err := conn.NewSession()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("new session failure, nest error: %v", err)
	}

	if err := session.RequestPty("vt220", int(ws.Rows), int(ws.Cols), ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		session.Close()
		conn.Close()
		return nil, fmt.Errorf("request pty failure, nest error: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		conn.Close()
		return nil, err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		conn.Close()
		return nil, err
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		conn.Close()
		return nil, err
	}

	if err := session.Shell(); err != nil {
		session.Close()
		conn.Close()
		return nil, err
	}
	return &Client{conn: conn, session: session, stdout: stdout, stderr: stderr, stdin: stdin}, nil
}

func setKeyboard(password string) func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
	return func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
		answers = make([]string, len(questions))
		for n := range questions {
			answers[n] = password
		}
		return answers, nil
	}
}

func (c *Client) ChangeWindow(ws *remote.WinSize) error {
	return c.session.WindowChange(int(ws.Rows), int(ws.Cols))
}

func (c *Client) Close() error {
	var eg []error
	if c.session != nil {
		if err := c.session.Close(); err != nil {
			eg = append(eg, err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			eg = append(eg, err)
		}
	}
	if len(eg) == 0 {
		return nil
	}
	return fmt.Errorf("%v", eg)
}

func (c *Client) Stdout() (io.Reader, error) {
	if c.session == nil {
		return nil, fmt.Errorf("panic: session is nil")
	}
	return c.stdout, nil
}

func (c *Client) Stdin() (io.WriteCloser, error) {
	if c.session == nil {
		return nil, fmt.Errorf("panic: session is nil")
	}
	return c.stdin, nil
}

func (c *Client) Stderr() (io.Reader, error) {
	if c.session == nil {
		return nil, fmt.Errorf("panic: session is nil")
	}
	return c.stderr, nil
}

func (c *Client) Wait() error {
	if c.session == nil {
		return fmt.Errorf("panic: session is nil")
	}
	return c.session.Wait()
}

func (c *Client) SetSignal(signal chan error) error {
	return nil
}
