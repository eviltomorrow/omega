package swap

import (
	"testing"

	"github.com/eviltomorrow/omega/pkg/plugins"
)

func TestGather(t *testing.T) {
	swap := &SwapStats{
		ps: plugins.NewSystemPS(),
	}
	metrics, err := swap.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range metrics {
		t.Logf("%s\r\n", m.String())
	}
}
