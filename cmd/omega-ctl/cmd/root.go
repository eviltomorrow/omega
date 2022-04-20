package cmd

import (
	"strconv"
	"time"

	"github.com/eviltomorrow/omega/pkg/self"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "omega-ctl",
	Short: "",
	Long:  "  \r\nomega-ctl is a tool for omega",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var (
	timeout         = 10 * time.Second
	addr            string
	Timeout         string
	EtcdKeyPrefix   = self.EtcdKeyPrefix
	EtcdEndpoints   []string
	MaxTimeoutLimit = 60 * time.Second
	MinTimeoutLimit = 5 * time.Second
)

func init() {
	root.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	root.AddCommand(service_root)
	root.AddCommand(omega_root)
	root.AddCommand(watchdog_root)
	root.AddCommand(hub_root)
}

func Execute() error {
	return root.Execute()
}

func setTimeout(s string) time.Duration {
	if Timeout != "" {
		d, err := time.ParseDuration(Timeout)
		if err == nil {
			timeout = d
		} else {
			i, err := strconv.Atoi(Timeout)
			if err == nil {
				timeout = time.Duration(i) * time.Second
			}
		}
	}
	if timeout > MaxTimeoutLimit {
		timeout = MaxTimeoutLimit
	}
	if timeout < MinTimeoutLimit {
		timeout = MinTimeoutLimit
	}
	return timeout
}
