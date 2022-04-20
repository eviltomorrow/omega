package hub

import (
	"log"
	"testing"

	"github.com/eviltomorrow/omega/pkg/self"
)

func TestPull(t *testing.T) {
	destroy, err := self.RegisterEtcd([]string{"127.0.0.1:2379"})
	if err != nil {
		log.Printf("[E] Register etcd failure, nest error: %v", err)
		return
	}
	defer destroy()

	if err := Pull("/tmp/omega", "latest"); err != nil {
		log.Fatal(err)
	}
}
