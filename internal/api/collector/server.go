package collector

import (
	"io"

	"github.com/eviltomorrow/omega/internal/api/collector/pb"
	"github.com/eviltomorrow/omega/pkg/zlog"
	"go.uber.org/zap"
)

type Server struct {
	pb.UnimplementedCollectorServer
}

func (s *Server) Push(ps pb.Collector_PushServer) error {
	for {
		metric, err := ps.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		zlog.Info("metric", zap.Any("data", metric))
	}
	return nil
}
