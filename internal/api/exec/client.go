package exec

import (
	"context"
	"fmt"

	"github.com/eviltomorrow/omega/internal/api/exec/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(target string) (pb.ExecClient, func(), error) {
	conn, err := grpc.DialContext(
		context.Background(),
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("dial %s failure, nest error: %v", target, err)
	}

	return pb.NewExecClient(conn), func() { conn.Close() }, nil
}
