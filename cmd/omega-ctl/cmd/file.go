package cmd

import (
	"log"

	"github.com/eviltomorrow/omega/internal/api/file"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var file_upload = &cobra.Command{
	Use:   "upload",
	Short: "upload file to omega",
	Long:  "  \r\file api(Upload)",
	Run: func(cmd *cobra.Command, args []string) {
		err := apiFileUpload()
		if err != nil {
			log.Printf("[E] Upload to omega failure, nest error: %v", err)
		} else {
			log.Printf("[%s]", color.BlueString("OK"))
		}
	},
}

var file_download = &cobra.Command{
	Use:   "download",
	Short: "download file from omega",
	Long:  "  \r\file api(Download)",
	Run: func(cmd *cobra.Command, args []string) {
		err := apiFileDownload()
		if err != nil {
			log.Printf("[E] Download from omega failure, nest error: %v", err)
		} else {
			log.Printf("[%s]", color.BlueString("OK"))
		}
	},
}

func apiFileUpload() error {
	return file.Upload(addr, local, remote)
}

func apiFileDownload() error {
	return file.Download(addr, local, remote)
}
