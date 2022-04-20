package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/eviltomorrow/omega/internal/api/watchdog"
	"github.com/eviltomorrow/omega/internal/api/watchdog/pb"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var watchdog_root = &cobra.Command{
	Use:   "watchdog",
	Short: "watchdog's api support",
	Long:  "  \r\nomega-ctl watchdog api support",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var watchdog_notify = &cobra.Command{
	Use:   "notify",
	Short: "handle omega startup or shutdown",
	Long:  "  \r\nwatchdog api(notify)",
	Run: func(cmd *cobra.Command, args []string) {
		pid, err := apiWatchdogNotify()
		if err != nil {
			log.Printf("[E] Notfiy signal failure, nest error: %v", err)
		} else {
			log.Printf("[I] Notify signal success, pid: %v", pid)
		}
	},
}

var watchdog_pull = &cobra.Command{
	Use:   "pull",
	Short: "pull image omega with specify tag",
	Long:  "  \r\nwatchdog api(pull)",
	Run: func(cmd *cobra.Command, args []string) {
		md5, err := apiWatchdogPull()
		if err != nil {
			log.Printf("[E] Pull image failure, nest error: %v", err)
		} else {
			log.Printf("[I] Success, md5: %v", md5)
		}
	},
}

var (
	sig, tag string
)

func init() {
	// notify
	watchdog_root.AddCommand(watchdog_notify)
	watchdog_notify.Flags().StringVar(&sig, "sig", "", "handle omega by watchdog[quit/up]")
	watchdog_notify.MarkFlagRequired("sig")
	watchdog_notify.Flags().StringVar(&addr, "addr", "", "wartchdog'service addr")
	watchdog_notify.MarkFlagRequired("addr")
	watchdog_notify.Flags().StringVar(&Timeout, "timeout", "10s", "watchdog's api timeout")

	// pull
	watchdog_root.AddCommand(watchdog_pull)
	watchdog_pull.Flags().StringVar(&addr, "addr", "", "wartchdog'service addr")
	watchdog_pull.MarkFlagRequired("addr")
	watchdog_pull.Flags().StringVar(&tag, "tag", "", "pull image with specify tag")
	watchdog_pull.MarkFlagRequired("tag")
}

func apiWatchdogNotify() (int32, error) {
	if sig == "" {
		return 0, fmt.Errorf("optional handle value[quit/up]")
	}
	if addr == "" || !strings.Contains(addr, ":") {
		return 0, fmt.Errorf("addr format error")
	}
	var timeout = setTimeout(Timeout)

	var signal = pb.Signal_UP
	if sig == "quit" {
		signal = pb.Signal_QUIT
	}

	client, close, err := watchdog.NewClient(addr)
	if err != nil {
		return 0, fmt.Errorf("create grpc client failure, nest error: %v", err)
	}
	defer close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := client.Notify(ctx, &pb.Signal{Signal: signal})
	if err != nil {
		return 0, err
	}
	return resp.Value, nil
}

func apiWatchdogPull() (string, error) {
	client, destroy, err := watchdog.NewClient(addr)
	if err != nil {
		return "", err
	}
	defer destroy()

	resp, err := client.Pull(context.Background(), &wrapperspb.StringValue{Value: tag})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}
