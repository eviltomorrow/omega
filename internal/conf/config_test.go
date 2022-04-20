package conf

import "testing"

func TestLoadConfig(t *testing.T) {
	var c = &Config{}
	if err := c.LoadFile("../../etc/omega.conf"); err != nil {
		t.Fatal(err)
	}
	t.Logf("conf: %v", c)
}
