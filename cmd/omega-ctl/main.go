package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eviltomorrow/omega/cmd/omega-ctl/cmd"
	"github.com/eviltomorrow/omega/internal/system"
)

const (
	envOmegaEndpoint   = "OMEGA_ENDPOINT"
	envOmegaMaxTimeout = "OMEGA_MAX_TIMEOUT"
	envOmegaMinTimeout = "OMEGA_MIN_TIMEOUT"
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
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	if err := setGlobalVars(); err != nil {
		log.Fatalf("[F] Set Global Var failure, nest error: %v", err)
	}
	if err := cmd.Execute(); err != nil {
		log.Fatalf("[F] Execute omege-ctl failure, nest error: %v", err)
	}
}

type logWriter struct {
}

func (writer logWriter) Write(buf []byte) (int, error) {
	return fmt.Printf("%s", string(buf))
}

func setGlobalVars() error {
	var ep = os.Getenv(envOmegaEndpoint)
	if ep == "" {
		return fmt.Errorf("you are not set environment variables[%s], eg. export %s=\"127.0.0.1:2379,192.168.xx.xx:2379\"", envOmegaEndpoint, envOmegaEndpoint)
	}
	var (
		epList    = strings.Split(ep, ",")
		endpoints = make([]string, 0, len(epList))
	)
	for _, endpoint := range epList {
		endpoint = strings.TrimSpace(endpoint)
		if endpoint == "" {
			continue
		}
		endpoints = append(endpoints, endpoint)
	}
	if len(endpoints) == 0 {
		return fmt.Errorf("not include valid endpoints, size is 0")
	}
	cmd.EtcdEndpoints = endpoints

	var (
		maxTimeout = os.Getenv(envOmegaMinTimeout)
		minTimeout = os.Getenv(envOmegaMinTimeout)
		d          time.Duration
		err        error
	)

	if maxTimeout != "" {
		d, err = time.ParseDuration(maxTimeout)
		if err != nil {
			return fmt.Errorf("parse [%s] failure, nest error: %v", envOmegaMaxTimeout, err)
		}
		cmd.MaxTimeoutLimit = d
	}

	if minTimeout != "" {
		d, err = time.ParseDuration(minTimeout)
		if err != nil {
			return fmt.Errorf("parse [%s] failure, nest error: %v", envOmegaMinTimeout, err)
		}
		cmd.MinTimeoutLimit = d
	}

	return nil
}
