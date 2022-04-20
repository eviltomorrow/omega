package bar

import (
	"fmt"
	"os"

	"github.com/creack/pty"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar"
)

func NewProgressbar(size int, mode string, desc string) (*progressbar.ProgressBar, chan int) {
	var (
		counter = make(chan int, 1)
	)

	var width = 100
	_, col, err := pty.Getsize(os.Stdout)
	if err == nil {
		width = col
		// md5
		if width > 32 {
			width = width - 32
		}
		// header
		if width > 28 {
			width = width - 28
		}
		// speed
		if width > 20 {
			width = width - 20
		}
		// blank
		if width > 10 {
			width = width - 10
		}
	}

	bar := progressbar.NewOptions(size,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(width),
		progressbar.OptionSetDescription(fmt.Sprintf("[green]%sing[reset] %s", mode, desc)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[white]=[reset]",
			SaucerHead:    "[white]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
		}),
	)

	go func() {
		for count := range counter {
			bar.Add(count)
		}
	}()
	return bar, counter
}
