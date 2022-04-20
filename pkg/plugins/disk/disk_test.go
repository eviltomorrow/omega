package disk

import (
	"testing"

	"github.com/eviltomorrow/omega/pkg/plugins"
)

func TestGather(t *testing.T) {
	disk := &DiskStats{
		ps: plugins.NewSystemPS(),
	}
	metrics, err := disk.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range metrics {
		t.Logf("%s\r\n", m.String())
	}
}
