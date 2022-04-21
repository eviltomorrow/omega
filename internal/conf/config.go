package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/eviltomorrow/omega/internal/system"
	"github.com/eviltomorrow/omega/pkg/zlog"
)

type Plugin map[string]interface{}
type Addr struct {
	IP string `toml:"ip" json:"ip"`
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func (p Plugin) Byte() []byte {
	buf, _ := json.Marshal(p)
	return buf
}

type Config struct {
	GrpcServerHost map[string]Addr   `toml:"grpc-server-host" json:"grpc-server-host"`
	Global         Global            `toml:"global" json:"global"`
	Log            Log               `toml:"log" json:"log"`
	Watchdog       Watchdog          `toml:"watchdog" json:"watchdog"`
	Agent          Agent             `toml:"agent" json:"agent"`
	Plugins        map[string]Plugin `toml:"plugins" json:"plugins"`
}

type Global struct {
	EtcdEndpoints  []string `toml:"etcd-endpoints" json:"etcd-endpoints"`
	GroupName      string   `toml:"group-name" json:"group-name"`
	GrpcServerPort int      `toml:"grpc-server-port" json:"grpc-server-port"`
}

type Log struct {
	DisableTimestamp bool   `json:"disable-timestamp" toml:"disable-timestamp"`
	Level            string `json:"level" toml:"level"`
	Format           string `json:"format" toml:"format"`
	Dir              string `json:"dir" toml:"dir"`
	MaxSize          int    `json:"maxsize" toml:"maxsize"`
}

type Watchdog struct {
	GrpcServerPort int `toml:"grpc-server-port" json:"grpc-server-port"`
}

type Agent struct {
	GrpcServerPort int      `toml:"grpc-server-port" json:"grpc-server-port"`
	Period         Duration `toml:"period" json:"period"`
}

type Collector struct {
	GrpcServerHost map[string]Addr `toml:"grpc-server-host" json:"grpc-server-host"`
	Global         Global          `toml:"global" json:"global"`
	Log            Log             `toml:"log" json:"log"`
}

func (c *Collector) LoadFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}

type Hub struct {
	GrpcServerHost map[string]Addr `toml:"grpc-server-host" json:"grpc-server-host"`
	Global         Global          `toml:"global" json:"global"`
	Log            Log             `toml:"log" json:"log"`
}

func (h *Hub) LoadFile(path string) error {
	_, err := toml.DecodeFile(path, h)
	return err
}

func (c *Config) LoadFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}

func (c *Config) String() string {
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("marshal conf failure, nest error: %v", err)
	}
	return string(buf)
}

func FindPath(baseDir, path string, suffix string) (string, error) {
	var possibleConf = []string{
		path,
		filepath.Join(baseDir, fmt.Sprintf("../etc/omega%s.conf", suffix)),
		filepath.Join(baseDir, fmt.Sprintf("./etc/omega%s.conf", suffix)),
		filepath.Join(baseDir, fmt.Sprintf("/etc/omega%s.conf", suffix)),
	}
	for _, path := range possibleConf {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			fp, err := filepath.Abs(path)
			if err == nil {
				return fp, nil
			}
			return path, nil
		}
	}
	return "", fmt.Errorf("not find omega-%s.conf, possible path: %v", suffix, possibleConf)
}

func SetupLog(log Log, fileName string) error {
	log.Dir = filepath.Join(system.RootDir, log.Dir)
	global, prop, err := zlog.InitLogger(&zlog.Config{
		Level:            log.Level,
		Format:           log.Format,
		DisableTimestamp: log.DisableTimestamp,
		File: zlog.FileLogConfig{
			Filename:   filepath.Join(log.Dir, fileName),
			MaxSize:    log.MaxSize,
			MaxDays:    30,
			MaxBackups: 30,
			Compress:   true,
		},
		DisableStacktrace:   true,
		DisableErrorVerbose: true,
	})
	if err != nil {
		return err
	}
	zlog.ReplaceGlobals(global, prop)
	return nil
}

var DefaultGlobalOmega = &Config{
	Log: Log{
		DisableTimestamp: false,
		Level:            "info",
		Format:           "text",
		Dir:              "../log",
		MaxSize:          20,
	},
	Global: Global{
		EtcdEndpoints: []string{
			"127.0.0.1:2379",
		},
		GroupName: "omega-default",
	},
	Watchdog: Watchdog{
		GrpcServerPort: 28500,
	},
	Agent: Agent{
		GrpcServerPort: 28501,
		Period: Duration{
			Duration: 60 * time.Second,
		},
	},
	Plugins: map[string]Plugin{
		"cpu": map[string]interface{}{
			"percpu":           false,
			"totalcpu":         true,
			"collect_cpu_time": false,
			"report_active":    true,
		},
		"disk": map[string]interface{}{
			"mount_points": nil,
			"ignore_fs":    nil,
		},
		"diskio": map[string]interface{}{
			"devices":            nil,
			"device_tags":        nil,
			"name_templates":     nil,
			"skip_serial_number": true,
		},
		"net": map[string]interface{}{
			"ignore_protocol_stats": false,
			"interfaces":            nil,
		},
		"processes": map[string]interface{}{
			"force_ps":   false,
			"force_proc": false,
		},
	},
}

var DefaultGlobalCollector = &Collector{
	Log: Log{
		DisableTimestamp: false,
		Level:            "info",
		Format:           "text",
		Dir:              "../log",
		MaxSize:          20,
	},
	Global: Global{
		EtcdEndpoints: []string{
			"127.0.0.1:2379",
		},
		GroupName:      "omega-default",
		GrpcServerPort: 30123,
	},
}

var DefaultGlobalHub = &Hub{
	Log: Log{
		DisableTimestamp: false,
		Level:            "info",
		Format:           "text",
		Dir:              "../log",
		MaxSize:          20,
	},
	Global: Global{
		EtcdEndpoints: []string{
			"127.0.0.1:2379",
		},
		GrpcServerPort: 30588,
	},
}
