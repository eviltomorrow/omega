package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/eviltomorrow/omega/internal/api/hub"
	"github.com/eviltomorrow/omega/internal/api/watchdog"
	"github.com/eviltomorrow/omega/internal/conf"
	server "github.com/eviltomorrow/omega/internal/server/omega-watchdog"
	"github.com/eviltomorrow/omega/internal/system"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/eviltomorrow/omega/pkg/lock"
	"github.com/eviltomorrow/omega/pkg/self"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "omega-watchdog",
	Short: "",
	Long:  "  \r\nomega-watchdog is a watchdog service",
	Run: func(cmd *cobra.Command, args []string) {
		if daemon {
			if err := self.RunDaemon("omega-watchdog", []string{"-c", cfgFile, "-p", pidFile}); err != nil {
				log.Fatalf("[F] Run omega-watchdog daemon failure, nest error: %v\r\n", err)
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
			filepath.Join(system.RootDir, "../var/run"),
			filepath.Join(system.RootDir, "../var/images"),
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

		destroy, err := self.RegisterEtcd(server.Endpoints)
		if err != nil {
			log.Printf("[E] Register etcd failure, nest error: %v", err)
			return
		}
		registerCleanFuncs(destroy)

		if err := pullImageLatest(); err != nil {
			log.Fatalf("[F] Pull latest image failure, nest error: %v\r\n", err)
		}

		if err := server.StartupGRPC(); err != nil {
			code = 1
			log.Printf("[F] Startup grpc server failure, nest error: %v\r\n", err)
			return
		}
		registerCleanFuncs(server.ShutdownGRPC)
		registerCleanFuncs(server.RevokeEtcdConn)

		go func() {
			var pidFile = filepath.Join(system.RootDir, "../var/run/omega.pid")
			alock, err := lock.CreateFileLock(pidFile)
			if err != nil {
				p, err := self.LoadChild(pidFile)
				if err != nil {
					log.Fatalf("[F] Run omega-wathdog failure, nest error: load child process failure, nest error: %v\r\n", err)
				}

				var (
					stop = make(chan struct{}, 1)
				)
				registerCleanFuncs(func() error {
					stop <- struct{}{}
					return nil
				})
				watchdog.Reload <- struct{}{}

			loop:
				for {
					select {
					case <-watchdog.Stop:
						if err := p.Signal(syscall.SIGQUIT); err != nil {
							log.Printf("[Panic] [Load]Process.Signal(SIGQUIT) failure, nest error: %v\r\n", err)
							p.Kill()
						} else {
							watchdog.Pid <- watchdog.PS{Pid: p.Pid}
							watchdog.Stop <- struct{}{}
							break loop
						}
					case <-stop:
						return
					}
				}
				<-watchdog.Reload

			} else {
				lock.DestroyFileLock(alock)
				watchdog.Reload <- struct{}{}
			}

			var writer = &lumberjack.Logger{
				Filename:   filepath.Join(system.RootDir, "../log/error.log"),
				MaxSize:    20,
				MaxBackups: 10,
				MaxAge:     28,
				Compress:   true,
			}

			for range watchdog.Reload {
				watchdog.Reload <- struct{}{}
				select {
				case <-watchdog.Stop:
				default:
				}

				cmd, err := self.RunChild("omega", []string{"-c", "omega.conf", "-p", pidFile}, writer)
				if err != nil {
					log.Printf("[Panic] [Reload] Run omega failure, nest error: %v\r\n", err)
					watchdog.Pid <- watchdog.PS{Err: err}
				} else {
					var (
						stop = make(chan struct{}, 1)
						sig  = make(chan error, 1)
					)

					// stop
					go func(c *exec.Cmd) {
						if c == nil || c.Process == nil {
							return
						}

						select {
						case <-watchdog.Stop:
							if err := c.Process.Signal(syscall.SIGQUIT); err != nil {
								log.Printf("[Panic] [Reload]Process.Signal(SIGQUIT) failure, nest error: %v\r\n", err)
								c.Process.Kill()
							}

							ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
							select {
							case <-ctx.Done():
								cancel()
								watchdog.Pid <- watchdog.PS{Pid: c.Process.Pid, Err: fmt.Errorf("Process.Signal(quit) Cost is more than 2s")}
							case err := <-sig:
								cancel()
								watchdog.Pid <- watchdog.PS{Pid: c.Process.Pid, Err: err}
							}

						case <-stop:

						}
					}(cmd)

					// start
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					go func(c *exec.Cmd) {
						if c == nil || c.Process == nil {
							return
						}
						select {
						case <-ctx.Done():
							cancel()
							watchdog.Pid <- watchdog.PS{Pid: c.Process.Pid}
							// if err := crontab.RegisterRebootWatchdog(); err != nil {
							// 	log.Printf("[W] Register watchdog to crontab failure, nest error: %v\r\n", err)
							// }

						case err := <-sig:
							cancel()
							if err != nil {
								stop <- struct{}{}
								close(stop)
								watchdog.Pid <- watchdog.PS{Err: err}
							} else {
								log.Printf("[Panic] [Reload] Cmd.Wait's Cost is more than 2s")
							}
						}
					}(cmd)
					sig <- cmd.Wait()
					close(sig)
					watchdog.Stop <- struct{}{}

				}
				<-watchdog.Reload
			}
		}()

		plock, err := lock.CreateFileLock(pidFile)
		if err != nil {
			log.Fatalf("[F] Create pid-file[var/run/omega-watchdog.pid] failure, nest error: %v\r\n", err)
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
	daemon        bool
)

func init() {
	root.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	root.Flags().StringVarP(&cfgFile, "config", "c", "omega-watchdog.conf", "omega-watchdog's config file")
	root.Flags().BoolVarP(&daemon, "daemon", "d", false, "omega-watchdog running in background")
	root.Flags().StringVarP(&pidFile, "pid", "p", "../var/run/omega-watchdog.pid", "omega-watchdog's pid file")
}

func Execute() {
	cobra.CheckErr(root.Execute())
}

func setupConfig() error {
	path, err := conf.FindPath(cfgFile, "")
	if err != nil {
		return fmt.Errorf("find config path failure, nest error: %v", err)
	}
	if err := DefaultGlobal.LoadFile(path); err != nil {
		return fmt.Errorf("load config with path[%s] failure, nest error: %v", path, err)
	}
	if err := conf.SetupLog(DefaultGlobal.Log, "watchdog.log"); err != nil {
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
	server.Port = DefaultGlobal.Watchdog.GrpcServerPort
	server.Endpoints = DefaultGlobal.Global.EtcdEndpoints
	server.Key = fmt.Sprintf("%s/omega-watchdog/%s", self.EtcdKeyPrefix, DefaultGlobal.Global.GroupName)
}

var mut sync.Mutex

func pullImageLatest() error {
	_, err := os.Stat(server.BinFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		} else {
			if err := hub.Pull(server.BinFile, "latest"); err != nil {
				return err
			}
		}
	}
	return nil
}

func registerCleanFuncs(f func() error) {
	mut.Lock()
	defer mut.Unlock()

	if f != nil {
		cleanFuncs = append(cleanFuncs, f)
	}
}

func runAndExitCleanFuncs(code int) {
	mut.Lock()
	defer mut.Unlock()

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
