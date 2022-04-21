package self

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

func LoadChild(pidFile string) (*os.Process, error) {
	buf, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return nil, err
	}
	pid, err := strconv.Atoi(string(buf))
	if err != nil {
		return nil, err
	}
	return os.FindProcess(pid)
}

func RunChild(name string, args []string, writer io.WriteCloser) (*exec.Cmd, error) {
	var data = make([]string, 0, len(args)+1)
	data = append(data, name)
	data = append(data, args...)
	var cmd = &exec.Cmd{
		Path:   name,
		Args:   data,
		Stdout: writer,
		Stderr: writer,
		Stdin:  os.Stdin,
	}
	return cmd, cmd.Start()
}

func RunDaemon(name string, args []string) error {
	var data = make([]string, 0, len(args)+1)
	data = append(data, name)
	data = append(data, args...)
	var cmd = &exec.Cmd{
		Path:   "/proc/self/exe",
		Args:   data,
		Stdout: nil,
		Stderr: nil,
		Stdin:  os.Stdin,
	}
	return cmd.Start()
}
