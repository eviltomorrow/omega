package net

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/eviltomorrow/omega"
	"github.com/eviltomorrow/omega/metric"
	"github.com/eviltomorrow/omega/pkg/plugins"
)

type NetIOStats struct {
	ps plugins.PS

	skipChecks          bool
	IgnoreProtocolStats bool     `json:"ignore_protocol_stats"`
	Interfaces          []string `json:"interfaces"`
}

func (n *NetIOStats) Description() string {
	return "Read metrics about network interface usage"
}

var netSampleConfig = `
  ## By default, telegraf gathers stats from any up interface (excluding loopback)
  ## Setting interfaces will tell it to gather these explicit interfaces,
  ## regardless of status.
  ##
  # interfaces = ["eth0"]
  ##
  ## On linux systems telegraf also collects protocol stats.
  ## Setting ignore_protocol_stats to true will skip reporting of protocol metrics.
  ##
  # ignore_protocol_stats = false
  ##
`

func (n *NetIOStats) SampleConfig() string {
	return netSampleConfig
}

func (n *NetIOStats) Gather() ([]omega.Metric, error) {
	netio, err := n.ps.NetIO()
	if err != nil {
		return nil, fmt.Errorf("get net io info failure, nest error: %v", err)
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("get list of interfaces failure, nest error: %v", err)
	}

	var now = time.Now()
	interfacesByName := map[string]net.Interface{}
	for _, iface := range interfaces {
		interfacesByName[iface.Name] = iface
	}

	var metrics = make([]omega.Metric, 0, 64)
	for _, io := range netio {
		if len(n.Interfaces) != 0 {
			var found bool

			// if n.filter.Match(io.Name) {
			// 	found = true
			// }

			if !found {
				continue
			}
		} else if !n.skipChecks {
			iface, ok := interfacesByName[io.Name]
			if !ok {
				continue
			}

			if iface.Flags&net.FlagLoopback == net.FlagLoopback {
				continue
			}

			if iface.Flags&net.FlagUp == 0 {
				continue
			}
		}

		tags := map[string]string{
			"interface": io.Name,
		}

		fields := map[string]interface{}{
			"bytes_sent":   io.BytesSent,
			"bytes_recv":   io.BytesRecv,
			"packets_sent": io.PacketsSent,
			"packets_recv": io.PacketsRecv,
			"err_in":       io.Errin,
			"err_out":      io.Errout,
			"drop_in":      io.Dropin,
			"drop_out":     io.Dropout,
		}
		metrics = append(metrics, metric.New("net", tags, fields, now, omega.Counter))
	}

	// Get system wide stats for different network protocols
	// (ignore these stats if the call fails)
	if !n.IgnoreProtocolStats {
		netprotos, _ := n.ps.NetProto()
		fields := make(map[string]interface{})
		for _, proto := range netprotos {
			for stat, value := range proto.Stats {
				name := fmt.Sprintf("%s_%s", strings.ToLower(proto.Protocol),
					strings.ToLower(stat))
				fields[name] = value
			}
		}
		tags := map[string]string{
			"interface": "all",
		}
		metrics = append(metrics, metric.New("net", tags, fields, now, omega.Unknown))
	}

	return metrics, nil
}

func (n *NetIOStats) Config(conf map[string]interface{}) error {
	return nil
}

func init() {
	plugins.Register("net", &NetIOStats{ps: plugins.NewSystemPS()})
}
