package cpu

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/eviltomorrow/omega"

	"github.com/eviltomorrow/omega/metric"
	"github.com/eviltomorrow/omega/pkg/plugins"

	cpuUtil "github.com/shirou/gopsutil/v3/cpu"
)

type CPUStats struct {
	ps        plugins.PS
	lastStats map[string]cpuUtil.TimesStat

	PerCPU         bool `json:"percpu"`
	TotalCPU       bool `json:"totalcpu"`
	CollectCPUTime bool `json:"collect_cpu_time"`
	ReportActive   bool `json:"report_active"`
}

func (c *CPUStats) Description() string {
	return "Read metrics about cpu usage"
}

var sampleConfig = `
  ## Whether to report per-cpu stats or not
  percpu = true
  ## Whether to report total system cpu stats or not
  totalcpu = true
  ## If true, collect raw CPU time metrics
  collect_cpu_time = false
  ## If true, compute and report the sum of all non-idle CPU states
  report_active = false
`

func (c *CPUStats) SampleConfig() string {
	return sampleConfig
}

func (c *CPUStats) Gather() ([]omega.Metric, error) {
	times, err := c.ps.CPUTimes(c.PerCPU, c.TotalCPU)
	if err != nil {
		return nil, fmt.Errorf("getting CPU info failure, nest error: %v", err)
	}

	var (
		now     = time.Now()
		metrics = make([]omega.Metric, 0, 64)
	)
	for _, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}

		total := totalCPUTime(cts)
		active := activeCPUTime(cts)

		if c.CollectCPUTime {
			fieldsC := map[string]interface{}{
				"time_user":       cts.User,
				"time_system":     cts.System,
				"time_idle":       cts.Idle,
				"time_nice":       cts.Nice,
				"time_iowait":     cts.Iowait,
				"time_irq":        cts.Irq,
				"time_softirq":    cts.Softirq,
				"time_steal":      cts.Steal,
				"time_guest":      cts.Guest,
				"time_guest_nice": cts.GuestNice,
			}
			if c.ReportActive {
				fieldsC["time_active"] = activeCPUTime(cts)
			}
			metrics = append(metrics, metric.New("cpu", tags, fieldsC, now, omega.Counter))
		}

		// Add in percentage
		if len(c.lastStats) == 0 {
			// If it's the 1st gather, can't get CPU Usage stats yet
			continue
		}

		lastCts, ok := c.lastStats[cts.CPU]
		if !ok {
			continue
		}
		lastTotal := totalCPUTime(lastCts)
		lastActive := activeCPUTime(lastCts)
		totalDelta := total - lastTotal

		if totalDelta < 0 {
			err = fmt.Errorf("current total CPU time is less than previous total CPU time")
			break
		}

		if totalDelta == 0 {
			continue
		}

		fieldsG := map[string]interface{}{
			"usage_user":       100 * (cts.User - lastCts.User - (cts.Guest - lastCts.Guest)) / totalDelta,
			"usage_system":     100 * (cts.System - lastCts.System) / totalDelta,
			"usage_idle":       100 * (cts.Idle - lastCts.Idle) / totalDelta,
			"usage_nice":       100 * (cts.Nice - lastCts.Nice - (cts.GuestNice - lastCts.GuestNice)) / totalDelta,
			"usage_iowait":     100 * (cts.Iowait - lastCts.Iowait) / totalDelta,
			"usage_irq":        100 * (cts.Irq - lastCts.Irq) / totalDelta,
			"usage_softirq":    100 * (cts.Softirq - lastCts.Softirq) / totalDelta,
			"usage_steal":      100 * (cts.Steal - lastCts.Steal) / totalDelta,
			"usage_guest":      100 * (cts.Guest - lastCts.Guest) / totalDelta,
			"usage_guest_nice": 100 * (cts.GuestNice - lastCts.GuestNice) / totalDelta,
		}
		if c.ReportActive {
			fieldsG["usage_active"] = 100 * (active - lastActive) / totalDelta
		}
		metrics = append(metrics, metric.New("cpu", tags, fieldsG, now, omega.Gauge))
	}

	c.lastStats = make(map[string]cpuUtil.TimesStat)
	for _, cts := range times {
		c.lastStats[cts.CPU] = cts
	}

	return metrics, err
}

func (c *CPUStats) String() string {
	buf, _ := json.Marshal(c)
	return string(buf)
}

func totalCPUTime(t cpuUtil.TimesStat) float64 {
	total := t.User + t.System + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal + t.Idle
	return total
}

func activeCPUTime(t cpuUtil.TimesStat) float64 {
	active := totalCPUTime(t) - t.Idle
	return active
}

func init() {
	ps := plugins.NewSystemPS()
	plugins.Register("cpu", &CPUStats{
		ps:             ps,
		PerCPU:         false,
		TotalCPU:       true,
		CollectCPUTime: false,
		ReportActive:   true,
	})
}
