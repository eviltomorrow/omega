package grpclb

import (
	"context"
	"fmt"
	"time"

	"github.com/eviltomorrow/omega/pkg/zlog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

var timeout = 5 * time.Second

func Register(service string, inner_ip, outer_ip string, port int, endpoints []string, ttl int64) (func(), error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
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
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, endpoint := range endpoints {
		_, err = client.Status(ctx, endpoint)
		if err != nil {
			return nil, err
		}
	}

	leaseResp, err := client.Grant(context.Background(), ttl)
	if err != nil {
		return nil, err
	}
	var leaseID = &leaseResp.ID

	key, value := fmt.Sprintf("/%s/%s:%d", service, inner_ip, port), fmt.Sprintf("%s:%d", outer_ip, port)
	_, err = client.Put(context.Background(), key, value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return nil, err
	}

	keepAlive, err := client.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return nil, err
	}

	var signal = make(chan struct{}, 1)
	go func() {
	keep:
		for {
			select {
			case <-client.Ctx().Done():
				return
			case k, ok := <-keepAlive:
				if !ok {
					break keep
				}
				if k != nil {
					_ = k
				}
			case <-signal:
				return
			}
		}

	release:
		zlog.Error("Etcd status is offline: register service retrying...", zap.String("key", key), zap.String("value", value))
		keepAlive, leaseID, err = registerRetry(client, key, value, ttl)
		if err != nil {
			zlog.Error("Retrying register service to etcd failure", zap.Error(err))
			time.Sleep(10 * time.Second)
			goto release
		}
		zlog.Info("Etcd status is online: register service complete", zap.String("key", key), zap.String("value", value))
		goto keep
	}()
	close := func() {
		signal <- struct{}{}
		if leaseID != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = client.Revoke(ctx, *leaseID)
		}
	}

	return close, nil
}

func registerRetry(client *clientv3.Client, key, value string, ttl int64) (<-chan *clientv3.LeaseKeepAliveResponse, *clientv3.LeaseID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return nil, nil, err
	}

	_, err = client.Put(context.Background(), key, value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return nil, nil, err
	}

	keepAlive, err := client.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return nil, nil, err
	}
	return keepAlive, &leaseResp.ID, nil
}
