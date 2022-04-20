package processes

import (
	"testing"
	"time"
)

func TestGather(t *testing.T) {
	g := &Processes{
		execPS:       execPS,
		readProcFile: readProcFile,
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
