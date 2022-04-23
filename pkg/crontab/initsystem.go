package crontab

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/eviltomorrow/omega/internal/system"
	"github.com/eviltomorrow/omega/pkg/exec"
)

func RegisterRebootWatchdog() error {
	var (
		c       = "crontab -l"
		rebootC = fmt.Sprintf("@reboot cd %s; sleep 3; ./omega-watchdog --daemon", system.RootDir)
	)

	stdout, stderr, err := exec.RunCmd(c, 5*time.Second)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if stdout != "" {
		buf.WriteString(stdout)
	}
	if stderr != "" {
		buf.WriteString(stderr)
	}

	var scanner = bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var text = scanner.Text()
		if strings.Contains(strings.TrimSpace(text), rebootC) {
			return nil
		}
	}

	c = fmt.Sprintf(`crontab -l > crontab.conf && echo "%s" >> crontab.conf && crontab crontab.conf && rm -f crontab.conf`, rebootC)
	_, _, err = exec.RunCmd(c, 5*time.Second)
	return err
}
