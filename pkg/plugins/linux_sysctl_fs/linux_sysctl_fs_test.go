package linux_sysctl_fs

import (
	"path"
	"testing"
	"time"
)

func TestGather(t *testing.T) {
	g := &SysctlFS{
		path: path.Join(GetHostProc(), "/sys/fs"),
	}
	for {
		metrics, err := g.Gather()
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range metrics {
			t.Logf("%s\r\n", m.String())
		}
		time.Sleep(1 * time.Second)
	}
}
