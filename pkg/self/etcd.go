package self

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/eviltomorrow/omega/pkg/grpclb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

func RegisterEtcd(endpoints []string) (func() error, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
		LogConfig: &zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.ErrorLevel),
			Development:      false,
			Encoding:         "json",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create etcd client failure, nest error: %v", err)
	}

	for i, endpoint := range endpoints {
		if err := statusEtcd(client, endpoint); err != nil {
			log.Printf("[W] Connect to etcd service failure, nest error: %v, endpoint: %v", err, endpoint)
			if i == len(endpoints)-1 {
				return nil, fmt.Errorf("connect to etcd service failure, nest error: no valid endpoint, endpoints: %v", strings.Join(endpoints, ","))
			}
		} else {
			break
		}
	}

	builder := &grpclb.Builder{
		Client: client,
	}
	resolver.Register(builder)

	return client.Close, nil
}

func statusEtcd(client *clientv3.Client, endpoint string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.Status(ctx, endpoint)
	return err
}
