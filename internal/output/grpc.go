package output

import (
	"context"
	"fmt"
	"time"

	"github.com/eviltomorrow/omega"
	"github.com/eviltomorrow/omega/internal/api/collector/pb"
	"github.com/eviltomorrow/omega/metric"
	"github.com/eviltomorrow/omega/pkg/self"
	"github.com/eviltomorrow/omega/pkg/zlog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	DefaultDialTimeout = 5 * time.Second
)

type GrpcClient struct {
	GroupName string
	// EtcdEndpoints []string

	// etcd        *clientv3.Client
	buffer      chan []omega.Metric
	closef      func()
	client      pb.CollectorClient
	pc          pb.Collector_PushClient
	destroyEtcd func() error
}

func NewGrpcClient(groupName string, endpoints []string) (omega.Output, error) {
	destroy, err := self.RegisterEtcd(endpoints)
	if err != nil {
		return nil, err
	}
	return &GrpcClient{GroupName: groupName, buffer: make(chan []omega.Metric, 128), destroyEtcd: destroy}, nil
}

func (gc *GrpcClient) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultDialTimeout)
	defer cancel()

	// target := fmt.Sprintf("etcd:///%s", gc.GroupName)
	target := fmt.Sprintf("etcd:///grpclb/service/omega-collector/%s", gc.GroupName)
	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("grpc dial target[%s] failure, nest error: %v", target, err)
	}

	gc.client = pb.NewCollectorClient(conn)
	gc.closef = func() {
		conn.Close()
	}

	pc, err := gc.client.Push(context.TODO())
	if err != nil {
		gc.closef()
		return err
	}
	gc.pc = pc

	return nil
}

func (gc *GrpcClient) Close() error {
	if gc.pc != nil {
		gc.pc.CloseSend()
	}
	if gc.closef != nil {
		gc.closef()
	}
	if gc.destroyEtcd != nil {
		gc.destroyEtcd()
	}
	return nil
}

func (gc *GrpcClient) Start() {
	go func() {
		for metrics := range gc.buffer {
			data, ok := metric.SwitchMetricsToMetricSet(metrics)
			if !ok {
				continue
			}

			if err := gc.pc.Send(data); err != nil {
				gc.pc.CloseSend()
				gc.closef()

				var num = 1
				for {
					time.Sleep(time.Duration(num) * time.Second)
					if err := gc.Connect(); err == nil {
						break
					}
					zlog.Warn("Prepare to reconnect to collector", zap.String("retry-times(cost)", fmt.Sprintf("%ds(+%.0fs)", num, DefaultDialTimeout.Seconds())))

					if num >= 2<<7 {
						num = 1
					} else {
						num = num * 2
					}

				}
			}
		}
	}()
}

func (gc *GrpcClient) WriteMetric(metrics []omega.Metric) error {
	gc.buffer <- metrics

	return nil
}

func (gc *GrpcClient) Stop() {
	close(gc.buffer)
}
