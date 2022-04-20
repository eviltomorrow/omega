package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func RunCmd(c string, timeout time.Duration) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var (
		eg  = make(chan error)
		cmd = exec.Command("/bin/sh", "-c", c)
	)
	defer func() {
		close(eg)
	}()
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("start execute cmd failure, nest error: %v", err)
	}

	go func() {
		eg <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Signal(syscall.SIGINT)
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-eg
		return stdout.String(), stderr.String(), fmt.Errorf("execute cmd timeout")

	case err := <-eg:
		return stdout.String(), stderr.String(), err
	}
}

func RunShell(path string, args []string, timeout time.Duration) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var c = fmt.Sprintf("bash %s %s", path, strings.Join(args, " "))
	var (
		eg  = make(chan error)
		cmd = exec.Command("/bin/sh", "-c", c)
	)
	defer func() {
		close(eg)
	}()
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("start execute shell failure, nest error: %v", err)
	}

	go func() {
		eg <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
		<-eg
		return stdout.String(), stderr.String(), fmt.Errorf("execute shell timeout")

	case err := <-eg:
		return stdout.String(), stderr.String(), err
	}
}
