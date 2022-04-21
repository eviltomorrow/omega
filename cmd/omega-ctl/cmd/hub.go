package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/eviltomorrow/omega/internal/api/hub"
	"github.com/eviltomorrow/omega/pkg/self"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var hub_root = &cobra.Command{
	Use:   "hub",
	Short: "hub's api support",
	Long:  "  \r\nomega-ctl hub api support",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var hub_push = &cobra.Command{
	Use:   "push",
	Short: "push omega image to hub",
	Long:  "  \r\nhub api(push)",
	Run: func(cmd *cobra.Command, args []string) {
		destroy, err := self.RegisterEtcd(EtcdEndpoints)
		if err != nil {
			log.Printf("[E] Register etcd failure, nest error: %v", err)
			return
		}
		defer destroy()

		md5, err := apiHubPush()
		if err != nil {
			log.Printf("[E] Push image failure, nest error: %v", err)
		} else {
			log.Printf(" | %s [%s]", md5, color.GreenString("success"))
		}
	},
}

var hub_list = &cobra.Command{
	Use:   "list",
	Short: "list omega images",
	Long:  "  \r\nhub api(list)",
	Run: func(cmd *cobra.Command, args []string) {
		destroy, err := self.RegisterEtcd(EtcdEndpoints)
		if err != nil {
			log.Printf("[E] Register etcd failure, nest error: %v", err)
			return
		}
		defer destroy()

		if err := apiHubList(); err != nil {
			log.Printf("[E] List images failure, nest error: %v", err)
		}
	},
}

var hub_del = &cobra.Command{
	Use:   "del",
	Short: "del omega image",
	Long:  "  \r\nhub api(del)",
	Run: func(cmd *cobra.Command, args []string) {
		destroy, err := self.RegisterEtcd(EtcdEndpoints)
		if err != nil {
			log.Printf("[E] Register etcd failure, nest error: %v", err)
			return
		}
		defer destroy()

		result, err := apiHubDel()
		if err != nil {
			log.Printf("[E] List images failure, nest error: %v", err)
		} else {
			log.Println(result)
		}
	},
}

var (
	releaseNote string
)

func init() {
	// push
	hub_root.AddCommand(hub_push)
	hub_push.Flags().StringVar(&releaseNote, "release_note", "", "release_note about omega image")
	hub_push.MarkFlagRequired("release_note")
	hub_push.Flags().StringVar(&local, "local", "", "local path about omega image")
	hub_push.MarkFlagRequired("local")

	// list
	hub_root.AddCommand(hub_list)

	// del
	hub_root.AddCommand(hub_del)
	hub_del.Flags().StringVar(&tag, "tag", "", "omega image tag")
	hub_del.MarkFlagRequired("tag")
}

func apiHubPush() (string, error) {
	return hub.Push(local, releaseNote)
}

func apiHubDel() (string, error) {
	client, destroy, err := hub.NewClient()
	if err != nil {
		return "", err
	}
	defer destroy()

	resp, err := client.Del(context.Background(), &wrapperspb.StringValue{Value: tag})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func apiHubList() error {
	client, destroy, err := hub.NewClient()
	if err != nil {
		return err
	}
	defer destroy()

	resp, err := client.List(context.Background(), &emptypb.Empty{})
	if err != nil {
		return err
	}
	if len(resp.Images) == 0 {
		log.Printf("Empty")
	} else {
		var (
			no   int
			data [][]string = make([][]string, 0, len(resp.Images))
		)
		for _, image := range resp.Images {
			no++
			var lines = make([]string, 0, 5)
			lines = append(lines, fmt.Sprintf("%d", no))
			lines = append(lines, image.Tag)
			lines = append(lines, image.Md5)
			lines = append(lines, image.CreateTime)
			lines = append(lines, image.ReleaseNotes)
			data = append(data, lines)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"No", "Tag", "MD5", "CreateTime", "Note"})
		for _, v := range data {
			table.Append(v)
		}
		table.Render()
	}
	return nil
}
