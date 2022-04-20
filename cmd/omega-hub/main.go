package main

import (
	"github.com/eviltomorrow/omega/cmd/omega-hub/cmd"
	"github.com/eviltomorrow/omega/internal/system"
)

var (
	MainVersion = "unknown"
	GitSha      = "unknown"
	GitTag      = "unknown"
	GitBranch   = "unknown"
	BuildTime   = "unknown"
)

func init() {
	system.MainVersion = MainVersion
	system.GitSha = GitSha
	system.GitTag = GitTag
	system.GitBranch = GitBranch
	system.BuildTime = BuildTime
}

func main() {
	cmd.Execute()
}
