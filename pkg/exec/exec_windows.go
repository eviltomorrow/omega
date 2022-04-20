//go:build windows

package exec

import (
	"log"
	"os/exec"
	"time"
)

// ExecuteCmd 执行 command
func RunCmd(c string, timeout time.Duration) (string, string, error) {
	return "", "", fmt.Errorf("not implement")
}

func RunShell(path string, timeout time.Duration) (string, string, error) {
	return "", "", fmt.Errorf("not implement")
}
