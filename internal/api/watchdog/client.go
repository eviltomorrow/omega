package watchdog

import (
	"context"
	"fmt"
	"time"

	"github.com/eviltomorrow/omega/internal/api/watchdog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	DefaultDialTimeout = 5 * time.Second
)

func NewClient(target string) (pb.WatchdogClient, func(), error) {
	conn, err := grpc.DialContext(
		context.Background(),
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("dial %s failure, nest error: %v", target, err)
	}

	return pb.NewWatchdogClient(conn), func() { conn.Close() }, nil
}
