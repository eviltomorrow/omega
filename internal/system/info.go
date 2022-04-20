package system

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/eviltomorrow/omega/pkg/tools"
)

var (
	RootDir     string
	Pid         = os.Getpid()
	LaunchTime  = time.Now()
	HostName    string
	OS          = runtime.GOOS
	Arch        = runtime.GOARCH
	RunningTime = func() string {
		return tools.FormatDuration(time.Since(LaunchTime))
	}
	IP string
)

func init() {
	var (
		path string
		err  error
	)
	path, err = os.Executable()
	if err != nil {
		panic(fmt.Errorf("panic: get Executable path failure, nest error: %v", err))
	}
	path, err = filepath.Abs(path)
	if err != nil {
		panic(fmt.Errorf("panic: abs RootDir failure, nest error: %v", err))
	}
	RootDir = filepath.Dir(path)

	name, err := os.Hostname()
	if err == nil {
		HostName = name
	}

	localIP, err := tools.GetLocalIP()
	if err == nil {
		IP = localIP
	}
}

func GetInfo() string {
	var data = make(map[string]interface{}, 8)
	data["RootDir"] = RootDir
	data["Pid"] = Pid
	data["LaunchTime"] = LaunchTime
	data["HostName"] = HostName
	data["OS"] = OS
	data["Arch"] = Arch
	data["RunningTime"] = RunningTime()
	data["IP"] = IP

	buf, _ := json.Marshal(data)
	return string(buf)
}
