package exec

import (
	"testing"
	"time"
)

func TestExecCmd(t *testing.T) {
	stdout, stderr, err := RunCmd("ls -l", 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("stdout: %v\r\n", stdout)
	t.Logf("stderr: %v\r\n", stderr)

	stdout, stderr, err = RunCmd("ping www.baidu.com -c 1", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("stdout: %v\r\n", stdout)
	t.Logf("stderr: %v\r\n", stderr)
}

func TestExecShell(t *testing.T) {
	stdout, stderr, err := RunShell("../../tests/exec/ls.sh", nil, 10*time.Second)
	if err != nil {
		t.Error(err)
	}
	t.Logf("stdout: %v\r\n", stdout)
	t.Logf("stderr: %v\r\n", stderr)

	// stdout, stderr, err = ExecuteShell("../../tests/exec/ping.sh", 2*time.Second)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Logf("stdout: %v\r\n", stdout)
	// t.Logf("stderr: %v\r\n", stderr)

	stdout, stderr, err = RunShell("../../tests/exec/echo.sh a", nil, 2*time.Second)
	if err != nil {
		t.Error(err)
	}
	t.Logf("stdout: %v\r\n", stdout)
	t.Logf("stderr: %v\r\n", stderr)
}
