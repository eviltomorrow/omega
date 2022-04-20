package ticker

import (
	"fmt"
	"testing"
	"time"
)

func TestTick(t *testing.T) {
	ticker := NewAlignedTicker(time.Now(), 2*time.Second, 0, 0)
	for t := range ticker.Elapsed() {
		fmt.Println(t)
	}
}
