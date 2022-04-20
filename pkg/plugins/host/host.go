package host

import (
	"github.com/eviltomorrow/omega"
)

type HostStats struct {
}

func (hs *HostStats) Description() string {
	return "Read metrics about host info"
}

func (hs *HostStats) SampleConfig() string { return "" }

func (hs *HostStats) Gather() ([]omega.Metric, error) {
	return nil, nil
}

func init() {
	// plugins.Register("host", &HostStats{})
}
