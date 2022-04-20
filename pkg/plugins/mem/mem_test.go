package mem

import (
	"runtime"
	"testing"

	"github.com/eviltomorrow/omega/pkg/plugins"
)

func TestGather(t *testing.T) {
	mem := &MemStats{
		ps:       plugins.NewSystemPS(),
		platform: runtime.GOOS,
	}
	metrics, err := mem.Gather()
	if err != nil {
		t.Fatal(err)
	}

	for _, m := range metrics {
		t.Logf("%s\r\n", m.String())
	}
}
