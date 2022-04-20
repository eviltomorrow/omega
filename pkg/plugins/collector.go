package plugins

import (
	"github.com/eviltomorrow/omega"
)

type Collector interface {
	SampleConfig() string
	Description() string
	Gather() ([]omega.Metric, error)
}

var Repository = map[string]Collector{}

func Register(name string, collector Collector) {
	Repository[name] = collector
}
