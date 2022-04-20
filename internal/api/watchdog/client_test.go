package watchdog

import (
	"context"
	"testing"

	"github.com/eviltomorrow/omega/internal/api/watchdog/pb"
)

func TestUP(t *testing.T) {
	client, close, err := NewClient("localhost:28501")
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	ps, err := client.Notify(context.Background(), &pb.Signal{
		Signal: pb.Signal_UP,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ps: %v\r\n", ps.Value)
}

func TestQuit(t *testing.T) {
	client, close, err := NewClient("localhost:28501")
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	ps, err := client.Notify(context.Background(), &pb.Signal{
		Signal: pb.Signal_QUIT,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ps: %v\r\n", ps.Value)
}
