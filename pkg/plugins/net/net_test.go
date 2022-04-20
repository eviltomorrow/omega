package net

import (
	"testing"

	"github.com/eviltomorrow/omega/pkg/plugins"
)

func TestNetioGather(t *testing.T) {
	net := &NetIOStats{
		ps:                  plugins.NewSystemPS(),
		IgnoreProtocolStats: true,
	}
	metrics, err := net.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range metrics {
		t.Logf("%s\r\n", m.String())
	}
}

func TestNetstatGather(t *testing.T) {
	net := &NetStats{
		ps: plugins.NewSystemPS(),
	}
	metrics, err := net.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range metrics {
		t.Logf("%s\r\n", m.String())
	}
}
