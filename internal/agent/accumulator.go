package agent

import (
	"bytes"

	"github.com/eviltomorrow/omega"
	"github.com/eviltomorrow/omega/pkg/zlog"
	"go.uber.org/zap"
)

var Accumulator omega.Accumulator

type accumulator struct {
	buffer chan []omega.Metric
	name   string
}

func NewAccumulator(name string, buffer chan []omega.Metric) omega.Accumulator {
	var ac = &accumulator{
		name:   name,
		buffer: buffer,
	}
	return ac
}

func (ac *accumulator) Name() string {
	return ac.name
}

func (ac *accumulator) AddMetric(metrics []omega.Metric) {
	select {
	case ac.buffer <- metrics:
	default:
		<-ac.buffer
		select {
		case ac.buffer <- metrics:
		default:
		}
		zlog.Warn("Accumulator's buffer is overflow, the metrics will be ignore", zap.String("metrics", withMetricsString(metrics)))
	}
}

func withMetricsString(metrics []omega.Metric) string {
	var buf bytes.Buffer
	for _, metric := range metrics {
		buf.WriteString(metric.String())
	}
	return buf.String()
}
