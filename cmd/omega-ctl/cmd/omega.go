package cmd

import (
	"context"
	"log"

	"github.com/eviltomorrow/omega/internal/api/agent"
	"github.com/eviltomorrow/omega/internal/api/exec"
	"github.com/eviltomorrow/omega/internal/api/exec/pb"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

var omega_root = &cobra.Command{
	Use:   "omega",
	Short: "omega's api support",
	Long:  "  \r\nomega-ctl omega api support",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var omega_version = &cobra.Command{
	Use:   "version",
	Short: "print omega version information",
	Long:  "  \r\vomega api(GetVersion)",
	Run: func(cmd *cobra.Command, args []string) {
		version, err := apiOmegaVersion()
		if err != nil {
			log.Printf("[E] Get omega version failure, nest error: %v", err)
		} else {
			log.Println(version)
		}
	},
}

var omega_system = &cobra.Command{
	Use:   "system",
	Short: "print omega system information",
	Long:  "  \r\vomega api(GetSystem)",
	Run: func(cmd *cobra.Command, args []string) {
		system, err := apiOmegaSystem()
		if err != nil {
			log.Printf("[E] Get omega system failure, nest error: %v", err)
		} else {
			log.Println(system)
		}
	},
}

var omega_exec = &cobra.Command{
	Use:   "exec",
	Short: "exec cmd with omega",
	Long:  "  \r\vomega api(Exec)",
	Run: func(cmd *cobra.Command, args []string) {
		result, err := apiOmegaExec()
		if err != nil {
			log.Printf("[E] Exec cmd failure, nest error: %v", err)
		} else {
			log.Println(result)
		}
	},
}

var omega_ping = &cobra.Command{
	Use:   "ping",
	Short: "ping omega",
	Long:  "  \r\vomega api(Ping)",
	Run: func(cmd *cobra.Command, args []string) {
		pong, err := apiOmegaPing()
		if err != nil {
			log.Printf("[E] Ping omega failure, nest error: %v", err)
		} else {
			log.Println(pong)
		}
	},
}

var (
	c             string
	local, remote string
)

func init() {
	// version
	omega_root.AddCommand(omega_version)
	omega_version.Flags().StringVar(&addr, "addr", "", "omgega'service addr")
	omega_version.MarkFlagRequired("addr")

	// system
	omega_root.AddCommand(omega_system)
	omega_system.Flags().StringVar(&addr, "addr", "", "omgega'service addr")
	omega_system.MarkFlagRequired("addr")

	// exec
	omega_root.AddCommand(omega_exec)
	omega_exec.Flags().StringVar(&addr, "addr", "", "omega'service addr")
	omega_exec.MarkFlagRequired("addr")
	omega_exec.Flags().StringVar(&c, "c", "", "command to run")
	omega_exec.MarkFlagRequired("c")
	omega_exec.Flags().StringVar(&Timeout, "timeout", "10s", "watchdog's api timeout")

	// ping
	omega_root.AddCommand(omega_ping)
	omega_ping.Flags().StringVar(&addr, "addr", "", "omega'service addr")
	omega_ping.MarkFlagRequired("addr")

	// upload
	omega_root.AddCommand(file_upload)
	file_upload.Flags().StringVar(&addr, "addr", "", "omega'service addr")
	file_upload.MarkFlagRequired("addr")
	file_upload.Flags().StringVar(&local, "local", "", "local file path")
	file_upload.MarkFlagRequired("local")
	file_upload.Flags().StringVar(&remote, "remote", "", "remote file path")
	file_upload.MarkFlagRequired("remote")

	// download
	omega_root.AddCommand(file_download)
	file_download.Flags().StringVar(&addr, "addr", "", "omega'service addr")
	file_download.MarkFlagRequired("addr")
	file_download.Flags().StringVar(&local, "local", "", "local file path")
	file_download.MarkFlagRequired("local")
	file_download.Flags().StringVar(&remote, "remote", "", "remote file path")
	file_download.MarkFlagRequired("remote")

	omega_root.AddCommand(terminal_root)
}

func apiOmegaVersion() (string, error) {
	client, destroy, err := agent.NewClient(addr)
	if err != nil {
		return "", err
	}
	defer destroy()

	resp, err := client.GetVersion(context.Background(), &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func apiOmegaSystem() (string, error) {
	client, destroy, err := agent.NewClient(addr)
	if err != nil {
		return "", err
	}
	defer destroy()

	resp, err := client.GetSystem(context.Background(), &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func apiOmegaExec() (string, error) {
	client, destroy, err := exec.NewClient(addr)
	if err != nil {
		return "", err
	}
	defer destroy()

	timeout := setTimeout(Timeout)
	resp, err := client.Run(context.Background(), &pb.C{
		Type:    pb.C_CMD,
		Text:    c,
		Timeout: int64(timeout),
	})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func apiOmegaPing() (string, error) {
	client, destroy, err := agent.NewClient(addr)
	if err != nil {
		return "", err
	}
	defer destroy()

	resp, err := client.Ping(context.Background(), &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}
