package cpu

import (
	"testing"
	"time"

	"github.com/eviltomorrow/omega/pkg/plugins"
)

func TestGather(t *testing.T) {
	cpu := &CPUStats{
		ps:             plugins.NewSystemPS(),
		PerCPU:         true,
		TotalCPU:       true,
		CollectCPUTime: true,
		ReportActive:   true,
	}
	for {
		metrics, err := cpu.Gather()
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range metrics {
			t.Logf("%s\r\n", m.String())
		}
		time.Sleep(1 * time.Second)
	}
}
