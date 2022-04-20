package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/eviltomorrow/omega/internal/api/terminal"
	"github.com/spf13/cobra"
)

var terminal_root = &cobra.Command{
	Use:   "terminal",
	Short: "terminal connect to omega",
	Long:  "  \r\termianl api(Exec)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := apiOmegaTerminal(); err != nil {
			log.Printf("Connect to omega terminal failure, nest error: %v", err)
		}
	},
}

var (
	mode string
)

func init() {
	// terminal
	terminal_root.Flags().StringVar(&mode, "mode", "local", "connect to omega mode[local]")
	terminal_root.MarkFlagRequired("mode")
	terminal_root.Flags().StringVar(&addr, "addr", "", "wartchdog'service addr")
	terminal_root.MarkFlagRequired("addr")

}

func apiOmegaTerminal() error {
	if addr == "" || !strings.Contains(addr, ":") {
		return fmt.Errorf("target is invalid")
	}

	var timeout = setTimeout(Timeout)
	switch mode {
	case "local":
		return terminal.NewLocal("/bin/bash", addr, timeout)
	// case "ssh":
	// 	if resource == "" {
	// 		return fmt.Errorf("resource is invalid")
	// 	}
	// 	var r = &terminal.Resource{}
	// 	if err := json.Unmarshal([]byte(resource), r); err != nil {
	// 		return err
	// 	}
	// 	return terminal.NewSSH("/bin/bash", target, timeout, r)
	default:
		return fmt.Errorf("not support mode[%s]", mode)
	}
}
