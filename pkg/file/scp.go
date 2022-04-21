package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type SCP struct {
	conn *ssh.Client
}

func NewSCP(username, password, host string, port int, pk []byte, timeout time.Duration) (*SCP, error) {
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
	return &SCP{conn: conn}, nil
}

func (s *SCP) Upload(localFile, remoteFile string) error {
	session, err := s.conn.NewSession()
	if err != nil {
		return fmt.Errorf("create session failure, nest error: %v", err)
	}
	defer session.Close()

	file, err := os.Open(localFile)
	if err != nil {
		return fmt.Errorf("open local file failure, nest error: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat local file failure, nest error: %v", err)
	}
	size := info.Size()

	go func(reader io.Reader, size int64, path string) {
		name := filepath.Base(path)
		stdin, _ := session.StdinPipe()
		fmt.Fprintln(stdin, "C0644", size, name)
		io.CopyN(stdin, reader, size)
		fmt.Fprint(stdin, "\x00")
		stdin.Close()
	}(file, size, localFile)

	dir := strings.Replace(filepath.Dir(remoteFile), "\\", "/", -1)
	cmd := fmt.Sprintf("/usr/bin/scp -qrt %s", dir)
	if err := session.Run(cmd); err != nil {
		return err
	}
	return nil
}

func (s *SCP) Run(c string, timeout time.Duration) ([]byte, []byte, error) {
	session, err := s.conn.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("create session failure, nest error: %v", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stderr = &stdout
	session.Stdout = &stderr

	if err := session.Start(c); err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var esig = make(chan error, 1)
	go func() {
		esig <- session.Wait()
		close(esig)
	}()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("exec timeout")
	case err := <-esig:
		return stdout.Bytes(), stderr.Bytes(), err
	}
}

func (s *SCP) Close() {
	if s != nil && s.conn != nil {
		s.conn.Close()
	}
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
