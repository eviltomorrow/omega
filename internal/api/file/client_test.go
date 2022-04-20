package file

import (
	"testing"
	"time"
)

func TestUpload(t *testing.T) {
	var now = time.Now()
	err := Upload("192.168.95.118:8090", "/home/shepard/下载/mysql-8.0.28-linux-glibc2.12-x86_64.tar.xz", "/root/app")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("cost: %v\r\n", time.Since(now))
}

func TestDownload(t *testing.T) {
	err := Download("localhost:8090", "/home/shepard/下载/mysql-8.0.28-linux-glibc2.12-x86_64.tar.xz", "/home/shepard/tmp/mysql-8.0.28-linux-glibc2.12-x86_64.tar.xz")
	if err != nil {
		t.Fatal(err)
	}
}
