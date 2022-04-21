package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/eviltomorrow/omega/internal/agent"
	"github.com/eviltomorrow/omega/internal/conf"
	server "github.com/eviltomorrow/omega/internal/server/omega"
	"github.com/eviltomorrow/omega/internal/system"
	"github.com/eviltomorrow/omega/pkg/lock"
	"github.com/eviltomorrow/omega/pkg/self"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "omega",
	Short: "",
	Long:  "  \r\nomega is a Host Monitoring Service",
	Run: func(cmd *cobra.Command, args []string) {
		for _, dir := range []string{
			filepath.Join(system.RootDir, "../var/cache"),
			filepath.Join(system.RootDir, "../var/run"),
			filepath.Join(system.RootDir, "../var/scripts"),
			filepath.Join(system.RootDir, "../log"),
		} {
			if err := initFolder(dir); err != nil {
				log.Fatalf("[F] Init folder failure, nest error: %v\r\n", err)
			}
		}

		var (
			code int
		)
		defer runAndExitCleanFuncs(code)

		logWriter := self.SetLog(filepath.Join(system.RootDir, "../log/error.log"))
		registerCleanFuncs(logWriter.Close)

		if err := setupConfig(); err != nil {
			code = 1
			log.Printf("[F] Setup config failure, nest error: %v", err)
			return
		}
		setupVars()

		instance, err := agent.NewAgent(DefaultGlobal)
		if err != nil {
			code = 1
			log.Printf("[F] New agent failure, nest error: %v\r\n", err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		registerCleanFuncs(func() error {
			cancel()
			return nil
		})

		var signal = make(chan struct{}, 1)
		go func() {
			if err := instance.Run(ctx, signal); err != nil {
				log.Printf("[F] Run agent failure, nest error: %v\r\n", err)
				logWriter.Close()
				os.Exit(1)
			}
		}()

		<-signal
		if err := server.StartupGRPC(); err != nil {
			code = 1
			log.Printf("[F] Startup grpc server failure, nest error: %v\r\n", err)
			return
		}
		registerCleanFuncs(server.ShutdownGRPC)
		registerCleanFuncs(server.RevokeEtcdConn)

		plock, err := lock.CreateFileLock(pidFile)
		if err != nil {
			code = 1
			log.Printf("[F] Create pid-file[%s] failure, nest error: %v", pidFile, err)
			return
		}
		registerCleanFuncs(func() error { return lock.DestroyFileLock(plock) })

		blockingUntilTermination()
	},
}

var (
	DefaultGlobal = conf.DefaultGlobalOmega
	cfgFile       = ""
	pidFile       = ""
	cleanFuncs    []func() error
)

func init() {
	root.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	root.Flags().StringVarP(&cfgFile, "config", "c", "omega.conf", "omega's config file")
	root.Flags().StringVarP(&pidFile, "pid", "p", "../var/run/omega.pid", "omega's pid file")
}

func Execute() {
	cobra.CheckErr(root.Execute())
}

func setupConfig() error {
	path, err := conf.FindPath(system.RootDir, cfgFile, "")
	if err != nil {
		return fmt.Errorf("find config path failure, nest error: %v", err)
	}
	if err := DefaultGlobal.LoadFile(path); err != nil {
		return fmt.Errorf("load config with path[%s] failure, nest error: %v", path, err)
	}
	if err := conf.SetupLog(DefaultGlobal.Log, "agent.log"); err != nil {
		return fmt.Errorf("setup zlog config failure, nest error: %v", err)
	}
	return nil
}

func setupVars() {
	var (
		addr conf.Addr
		ok   bool
	)
	addr, ok = DefaultGlobal.GrpcServerHost["inner_ip"]
	if ok {
		server.InnerIP = addr.IP
	}
	addr, ok = DefaultGlobal.GrpcServerHost["outer_ip"]
	if ok {
		server.OuterIP = addr.IP
	}
	server.Port = DefaultGlobal.Agent.GrpcServerPort
	server.Endpoints = DefaultGlobal.Global.EtcdEndpoints
	server.Key = fmt.Sprintf("%s/omega/%s", self.EtcdKeyPrefix, DefaultGlobal.Global.GroupName)
}

func registerCleanFuncs(f func() error) {
	if f != nil {
		cleanFuncs = append(cleanFuncs, f)
	}
}

func runAndExitCleanFuncs(code int) {
	for i := len(cleanFuncs) - 1; i >= 0; i-- {
		f := cleanFuncs[i]
		if f != nil {
			f()
		}
	}
	if code != 0 {
		os.Exit(code)
	}
}

func blockingUntilTermination() {
	var ch = make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	switch <-ch {
	case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
	case syscall.SIGUSR1:
	case syscall.SIGUSR2:
	default:
	}
}

func initFolder(dir string) error {
	fi, err := os.Stat(dir)
	if err == nil && !fi.IsDir() {
		return fmt.Errorf("exist same name file")
	}
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return err
}
