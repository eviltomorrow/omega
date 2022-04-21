package self

import (
	"io"
	"log"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	writer io.WriteCloser
	onece  sync.Once
)

func SetLog(path string) io.WriteCloser {
	onece.Do(func() {
		writer = &lumberjack.Logger{
			Filename:   path,
			MaxSize:    20,
			MaxBackups: 10,
			MaxAge:     28,
			Compress:   true,
		}
		log.SetOutput(writer)
	})
	return writer
}
