package agent

import (
	"context"
	"testing"

	"github.com/eviltomorrow/omega/internal/conf"
	_ "github.com/eviltomorrow/omega/pkg/plugins/cpu"
	"github.com/stretchr/testify/assert"
)

func TestConfigPlugins(t *testing.T) {
	judge := assert.New(t)
	var (
		c   = &conf.Config{}
		err = c.LoadFile("../../tests/conf/omega.conf")
	)
	judge.Nil(err)
	judge.Nil(ConfigPlugins(c))
}

func TestAgentRun(t *testing.T) {
	judge := assert.New(t)

	var (
		c   = &conf.Config{}
		err = c.LoadFile("../../tests/conf/omega.conf")
	)
	judge.Nil(err)

	agent, err := NewAgent(c)
	judge.Nil(err)

	err = agent.Run(context.Background(), nil)
	judge.Nil(err)
}
