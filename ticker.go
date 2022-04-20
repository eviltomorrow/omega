package omega

import "time"

type Ticker interface {
	Elapsed() <-chan time.Time
	Stop()
}
