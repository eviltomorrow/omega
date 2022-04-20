package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/eviltomorrow/omega/pkg/self"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

var service_root = &cobra.Command{
	Use:   "service",
	Short: "service's api support",
	Long:  "  \r\nomega-ctl service api support",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var service_list = &cobra.Command{
	Use:   "list",
	Short: "list running service",
	Long:  "  \r\nlist running service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := apiServiceList(""); err != nil {
			log.Printf("[F] List service failure, nest error: %v", err)
		}
	},
}

func init() {
	service_root.AddCommand(service_list)
}

func apiServiceList(service string) error {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   EtcdEndpoints,
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
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for i, endpoint := range EtcdEndpoints {
		_, err = client.Status(ctx, endpoint)
		if err != nil {
			log.Printf("[W] Connect to etcd service failure, nest error: %v, endpoint: %v", err, endpoint)
			if i == len(EtcdEndpoints)-1 {
				return fmt.Errorf("connect to etcd service failure, nest error: no valid endpoint, endpoints: %v", EtcdEndpoints)
			}
		} else {
			break
		}
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), timeout)
	defer cancel1()

	var key = "/" + self.EtcdKeyPrefix
	if service != "" {
		key += "/" + service + "/"
	}

	getResp, err := client.Get(ctx1, key, clientv3.WithPrefix())
	if err != nil {
		log.Printf("[F] Get key[%s] with prefix failure, nest error: %v", key, err)
	} else {
		if len(getResp.Kvs) == 0 {
			log.Printf("Empty")
		} else {
			var (
				no   int
				data [][]string = make([][]string, 0, len(getResp.Kvs))
			)
			for _, kv := range getResp.Kvs {
				no++
				var lines = make([]string, 0, 5)
				// no
				lines = append(lines, fmt.Sprintf("%d", no))

				// service
				k := string(kv.Key)
				l := ""
				if k != "" {
					if service != "" {
						lines = append(lines, service)
						l = strings.TrimPrefix(k, key)
					} else {
						l = strings.TrimPrefix(k, key)
						l = strings.TrimPrefix(l, "/")
						idx := strings.Index(l, "/")
						if idx != -1 {
							lines = append(lines, l[:idx])
							l = l[idx:]
						} else {
							lines = append(lines, "unknown")
						}
					}
				} else {
					lines = append(lines, "unknown")
				}

				// group
				if l != "" {
					l = strings.TrimPrefix(l, "/")
					idx := strings.Index(l, "/")
					if idx != -1 {
						lines = append(lines, l[:idx])
					} else {
						lines = append(lines, "unknown")
					}
				} else {
					lines = append(lines, "unknown")
				}
				lines = append(lines, string(kv.Key), string(kv.Value))
				data = append(data, lines)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"No", "Service", "Group", "Key", "value"})
			for _, v := range data {
				table.Append(v)
			}
			table.Render()
		}
	}
	return nil
}
