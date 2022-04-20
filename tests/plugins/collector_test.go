package plugins

import (
	"fmt"
	"testing"

	"github.com/eviltomorrow/omega/pkg/plugins"
	_ "github.com/eviltomorrow/omega/pkg/plugins/cpu"
)

func TestGather(t *testing.T) {
	for k, v := range plugins.Repository {
		fmt.Println(k)
		metrics, err := v.Gather()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(metrics)
	}
}

func BenchmarkGather(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range plugins.Repository {
			v.Gather()
		}
	}
}
