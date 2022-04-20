package net

import (
	"fmt"
	"syscall"
	"time"

	"github.com/eviltomorrow/omega"
	"github.com/eviltomorrow/omega/metric"
	"github.com/eviltomorrow/omega/pkg/plugins"
)

type NetStats struct {
	ps plugins.PS
}

func (ns *NetStats) Description() string {
	return "Read TCP metrics such as established, time wait and sockets counts."
}

var tcpstatSampleConfig = ""

func (ns *NetStats) SampleConfig() string {
	return tcpstatSampleConfig
}

func (ns *NetStats) Gather() ([]omega.Metric, error) {
	netconns, err := ns.ps.NetConnections()
	if err != nil {
		return nil, fmt.Errorf("getting net connections info failure, nest error: %v", err)
	}

	var metrics = make([]omega.Metric, 0, 64)
	var now = time.Now()
	counts := make(map[string]int)
	counts["UDP"] = 0

	// TODO: add family to tags or else
	tags := map[string]string{}
	for _, netcon := range netconns {
		if netcon.Type == syscall.SOCK_DGRAM {
			counts["UDP"]++
			continue // UDP has no status
		}
		c, ok := counts[netcon.Status]
		if !ok {
			counts[netcon.Status] = 0
		}
		counts[netcon.Status] = c + 1
	}

	fields := map[string]interface{}{
		"tcp_established": counts["ESTABLISHED"],
		"tcp_syn_sent":    counts["SYN_SENT"],
		"tcp_syn_recv":    counts["SYN_RECV"],
		"tcp_fin_wait1":   counts["FIN_WAIT1"],
		"tcp_fin_wait2":   counts["FIN_WAIT2"],
		"tcp_time_wait":   counts["TIME_WAIT"],
		"tcp_close":       counts["CLOSE"],
		"tcp_close_wait":  counts["CLOSE_WAIT"],
		"tcp_last_ack":    counts["LAST_ACK"],
		"tcp_listen":      counts["LISTEN"],
		"tcp_closing":     counts["CLOSING"],
		"tcp_none":        counts["NONE"],
		"udp_socket":      counts["UDP"],
	}
	metrics = append(metrics, metric.New("netstat", tags, fields, now, omega.Gauge))
	return metrics, nil
}

func (n *NetStats) Config(conf map[string]interface{}) error {
	return nil
}

func init() {
	plugins.Register("netstat", &NetStats{ps: plugins.NewSystemPS()})
}
