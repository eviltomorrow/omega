package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/eviltomorrow/omega/internal/conf"
	server "github.com/eviltomorrow/omega/internal/server/omega-hub"
	"github.com/eviltomorrow/omega/internal/system"
	"github.com/eviltomorrow/omega/pkg/lock"
	"github.com/eviltomorrow/omega/pkg/self"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "omega-hub",
	Short: "",
	Long:  "  \r\nomega-hub is a Hub Server",
	Run: func(cmd *cobra.Command, args []string) {
		if daemon {
			if err := self.RunDaemon("omega-hub", []string{"-c", cfgPath}); err != nil {
				log.Fatalf("[F] Run omega-hub daemon failure, nest error: %v\r\n", err)
			}
			os.Exit(0)
		}

		var (
			code int
		)
		defer runAndExitCleanFuncs(code)

		self.SetLog(filepath.Join(system.RootDir, "../log/error.log"))
		registerCleanFuncs(self.CloseLog)

		for _, dir := range []string{
			filepath.Join(system.RootDir, "../var/images"),
			filepath.Join(system.RootDir, "../var/run"),
			filepath.Join(system.RootDir, "../log"),
		} {
			if err := initFolder(dir); err != nil {
				log.Fatalf("[F] Init folder failure, nest error: %v\r\n", err)
			}
		}

		if err := setupConfig(); err != nil {
			log.Fatalf("[F] Setup config failure, nest error: %v\r\n", err)
		}
		setupVars()

		if err := server.StartupGRPC(); err != nil {
			code = 1
			log.Printf("[F] Startup grpc server failure, nest error: %v\r\n", err)
			return
		}
		registerCleanFuncs(server.ShutdownGRPC)
		registerCleanFuncs(server.RevokeEtcdConn)

		var pidFile = filepath.Join(system.RootDir, "../var/run/omega-hub.pid")
		plock, err := lock.CreateFileLock(pidFile)
		if err != nil {
			log.Fatalf("[F] Create pid-file[var/run/omega-hub.pid] failure, nest error: %v\r\n", err)
		}
		registerCleanFuncs(func() error { return lock.DestroyFileLock(plock) })

		blockingUntilTermination()
	},
}

var (
	DefaultGlobal = conf.DefaultGlobalHub
	cfgPath       = ""
	cleanFuncs    []func() error
	daemon        bool
)

func init() {
	root.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	root.Flags().StringVarP(&cfgPath, "config", "c", "omega-hub.conf", "omega-hub's config file")
	root.Flags().BoolVarP(&daemon, "daemon", "d", false, "omega-hub running in background")
}

func Execute() {
	cobra.CheckErr(root.Execute())
}

func setupConfig() error {
	path, err := conf.FindPath(cfgPath, "-hub")
	if err != nil {
		return fmt.Errorf("find config path failure, nest error: %v", err)
	}
	if err := DefaultGlobal.LoadFile(path); err != nil {
		return fmt.Errorf("load config with path[%s] failure, nest error: %v", path, err)
	}
	if err := conf.SetupLog(DefaultGlobal.Log, "hub.log"); err != nil {
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
	server.Port = DefaultGlobal.Global.GrpcServerPort
	server.Endpoints = DefaultGlobal.Global.EtcdEndpoints
	server.Key = fmt.Sprintf("%s/omega-hub", self.EtcdKeyPrefix)
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
