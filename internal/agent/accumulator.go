package agent

import (
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

func (ac *accumulator) AddMetric(metric []omega.Metric) {
	select {
	case ac.buffer <- metric:
	default:
		zlog.Warn("Accumulator's buffer is overflow, the metric will be ignore", zap.Any("metric", metric))
	}
}
