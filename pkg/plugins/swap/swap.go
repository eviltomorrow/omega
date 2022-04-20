package swap

import (
	"fmt"
	"time"

	"github.com/eviltomorrow/omega"
	"github.com/eviltomorrow/omega/metric"
	"github.com/eviltomorrow/omega/pkg/plugins"
)

type SwapStats struct {
	ps plugins.PS
}

func (ss *SwapStats) Description() string {
	return "Read metrics about swap memory usage"
}

func (ss *SwapStats) SampleConfig() string { return "" }

func (ss *SwapStats) Gather() ([]omega.Metric, error) {
	swap, err := ss.ps.SwapStat()
	if err != nil {
		return nil, fmt.Errorf("getting swap memory info failure, nest error: %v", err)
	}

	var (
		now     = time.Now()
		metrics = make([]omega.Metric, 0, 64)
	)

	fieldsG := map[string]interface{}{
		"total":        swap.Total,
		"used":         swap.Used,
		"free":         swap.Free,
		"used_percent": swap.UsedPercent,
	}
	fieldsC := map[string]interface{}{
		"in":  swap.Sin,
		"out": swap.Sout,
	}

	metrics = append(metrics, metric.New("swap", nil, fieldsG, now, omega.Gauge))
	metrics = append(metrics, metric.New("swap", nil, fieldsC, now, omega.Counter))
	return metrics, nil
}

func (s *SwapStats) Config(conf map[string]interface{}) error {
	return nil
}

func init() {
	ps := plugins.NewSystemPS()
	plugins.Register("swap", &SwapStats{
		ps: ps,
	})
}
