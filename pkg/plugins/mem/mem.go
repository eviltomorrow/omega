package mem

import (
	"fmt"
	"runtime"
	"time"

	"github.com/eviltomorrow/omega"
	"github.com/eviltomorrow/omega/metric"
	"github.com/eviltomorrow/omega/pkg/plugins"
)

type MemStats struct {
	ps       plugins.PS
	platform string
}

func (ms *MemStats) Description() string {
	return "Read metrics about memory usage"
}

func (ms *MemStats) SampleConfig() string { return "" }

func (ms *MemStats) Gather() ([]omega.Metric, error) {
	vm, err := ms.ps.VMStat()
	if err != nil {
		return nil, fmt.Errorf("get virtual memory info failure, nest error: %v", err)
	}
	now := time.Now()

	var metrics = make([]omega.Metric, 0, 1)

	fields := map[string]interface{}{
		"total":             vm.Total,
		"available":         vm.Available,
		"used":              vm.Used,
		"used_percent":      100 * float64(vm.Used) / float64(vm.Total),
		"available_percent": 100 * float64(vm.Available) / float64(vm.Total),
	}

	switch ms.platform {
	case "darwin":
		fields["active"] = vm.Active
		fields["free"] = vm.Free
		fields["inactive"] = vm.Inactive
		fields["wired"] = vm.Wired
	case "openbsd":
		fields["active"] = vm.Active
		fields["cached"] = vm.Cached
		fields["free"] = vm.Free
		fields["inactive"] = vm.Inactive
		fields["wired"] = vm.Wired
	case "freebsd":
		fields["active"] = vm.Active
		fields["buffered"] = vm.Buffers
		fields["cached"] = vm.Cached
		fields["free"] = vm.Free
		fields["inactive"] = vm.Inactive
		fields["laundry"] = vm.Laundry
		fields["wired"] = vm.Wired
	case "linux":
		fields["active"] = vm.Active
		fields["buffered"] = vm.Buffers
		fields["cached"] = vm.Cached
		fields["commit_limit"] = vm.CommitLimit
		fields["committed_as"] = vm.CommittedAS
		fields["dirty"] = vm.Dirty
		fields["free"] = vm.Free
		fields["high_free"] = vm.HighFree
		fields["high_total"] = vm.HighTotal
		fields["huge_pages_free"] = vm.HugePagesFree
		fields["huge_page_size"] = vm.HugePageSize
		fields["huge_pages_total"] = vm.HugePagesTotal
		fields["inactive"] = vm.Inactive
		fields["low_free"] = vm.LowFree
		fields["low_total"] = vm.LowTotal
		fields["mapped"] = vm.Mapped
		fields["page_tables"] = vm.PageTables
		fields["shared"] = vm.Shared
		fields["slab"] = vm.Slab
		fields["sreclaimable"] = vm.Sreclaimable
		fields["sunreclaim"] = vm.Sunreclaim
		fields["swap_cached"] = vm.SwapCached
		fields["swap_free"] = vm.SwapFree
		fields["swap_total"] = vm.SwapTotal
		fields["vmalloc_chunk"] = vm.VmallocChunk
		fields["vmalloc_total"] = vm.VmallocTotal
		fields["vmalloc_used"] = vm.VmallocUsed
		fields["write_back_tmp"] = vm.WriteBackTmp
		fields["write_back"] = vm.WriteBack
	}

	metrics = append(metrics, metric.New("mem", nil, fields, now, omega.Gauge))

	return metrics, nil
}

func (m *MemStats) Config(conf map[string]interface{}) error {
	return nil
}

func init() {
	ps := plugins.NewSystemPS()
	plugins.Register("mem", &MemStats{
		ps:       ps,
		platform: runtime.GOOS,
	})
}
