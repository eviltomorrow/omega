package system

import (
	"encoding/json"
	"runtime"
)

var (
	MainVersion string
	GoVersion   = runtime.Version()
	GoOSArch    = runtime.GOOS + "/" + runtime.GOARCH
	GitSha      string
	GitTag      string
	GitBranch   string
	BuildTime   string
)

func GetVersion() string {
	var data = make(map[string]string, 7)
	data["Current Version"] = MainVersion
	data["Go Version"] = GoVersion
	data["OS Arch"] = GoOSArch
	data["Git Sha"] = GitSha
	data["Git Tag"] = GitTag
	data["Git Branch"] = GitBranch
	data["Build Time"] = BuildTime

	buf, _ := json.Marshal(data)
	return string(buf)
}
