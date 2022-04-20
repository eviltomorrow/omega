package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/eviltomorrow/omega/internal/conf"
	"github.com/spf13/cobra"
)

var config_root = &cobra.Command{
	Use:   "config",
	Short: "Print version about omega-watchdog",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(apiWatchdogConfig())
	},
}

var (
	inner_ip, outer_ip string
	endpoints          string
	group              string
)

func init() {
	root.AddCommand(config_root)
	config_root.Flags().StringVar(&inner_ip, "inner_ip", "", "inner_ip propertites")
	config_root.Flags().StringVar(&outer_ip, "outer_ip", "", "outer_ip propertites")
	config_root.Flags().StringVar(&endpoints, "endpoints", "127.0.0.1:2379", "endpoints propertites")
	config_root.Flags().StringVar(&group, "group", "omega-default", "group propertites")
}

func apiWatchdogConfig() string {
	var config = &conf.Config{
		Global: conf.Global{
			EtcdEndpoints: []string{
				"127.0.0.1:2379",
			},
			GroupName: "omega-default",
		},
		Watchdog: conf.Watchdog{
			GrpcServerPort: 28500,
		},
		Agent: conf.Agent{
			GrpcServerPort: 28501,
			Period: conf.Duration{
				Duration: 60 * time.Second,
			},
		},
	}

	var addrs = map[string]conf.Addr{}
	if inner_ip != "" {
		addrs["inner_ip"] = conf.Addr{
			IP: inner_ip,
		}
	}
	if outer_ip != "" {
		addrs["outer_ip"] = conf.Addr{
			IP: outer_ip,
		}
	}
	var attrs = strings.Split(endpoints, ",")
	config.Global.EtcdEndpoints = attrs
	config.Global.GroupName = group
	if len(addrs) != 0 {
		config.GrpcServerHost = addrs
	}

	var buf bytes.Buffer
	if len(config.GrpcServerHost) != 0 {
		buf.WriteString("[grpc-server-host]\n")

		addr, ok := addrs["inner_ip"]
		if ok {
			buf.WriteString("    [grpc-server-host.inner_ip]\n")
			buf.WriteString(fmt.Sprintf("    ip = \"%s\"\n\n", addr.IP))
		}
		addr, ok = addrs["outer_ip"]
		if ok {
			buf.WriteString("    [grpc-server-host.outer_ip]\n")
			buf.WriteString(fmt.Sprintf("    ip = \"%s\"\n\n", addr.IP))
		}
	}

	buf.WriteString("[global]\n")
	buf.WriteString("etcd-endpoints = [\n")

	for _, attr := range attrs {
		buf.WriteString(fmt.Sprintf("    \"%s\",\n", attr))
	}
	buf.WriteString("]\n")
	buf.WriteString(fmt.Sprintf("group-name = \"%s\"\n", group))

	buf.WriteString("\n[watchdog]\n")
	buf.WriteString("grpc-server-port = 28500\n")

	buf.WriteString("\n[agent]\n")
	buf.WriteString("grpc-server-port = 28501\n")
	buf.WriteString("period = \"60s\"\n")
	return buf.String()
}
