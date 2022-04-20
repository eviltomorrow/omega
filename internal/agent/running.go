package agent

import (
	"sync"

	"github.com/eviltomorrow/omega"
	"github.com/eviltomorrow/omega/pkg/zlog"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var RunningOutputPool = &RunningOutput{
	outputs: make(map[string]omega.Output, 8),
	buffer:  make(chan []omega.Metric, 128),
}

type RunningOutput struct {
	mut     sync.Mutex
	outputs map[string]omega.Output

	buffer chan []omega.Metric
}

func (ro *RunningOutput) RegisterOutput(output omega.Output) (func(), error) {
	ro.mut.Lock()
	defer ro.mut.Unlock()

	uid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	var id = uid.String()
	ro.outputs[id] = output
	return func() {
		ro.mut.Lock()
		defer ro.mut.Unlock()

		delete(ro.outputs, id)
	}, nil
}

func (ro *RunningOutput) Buffer() chan []omega.Metric {
	return ro.buffer
}

func (ro *RunningOutput) Start() error {
	ro.mut.Lock()
	defer ro.mut.Unlock()

	for _, output := range ro.outputs {
		if err := output.Connect(); err != nil {
			return err
		}
		if running, ok := output.(omega.Running); ok {
			running.Start()
		}
	}

	go func() {
		for metrics := range ro.buffer {
			var outputs = ro.Range()
			for _, output := range outputs {
				if err := output.WriteMetric(metrics); err != nil {
					zlog.Error("Write metrics failure", zap.Error(err), zap.Any("metrics", metrics))
				}
			}
		}
	}()
	return nil
}

func (ro *RunningOutput) Range() []omega.Output {
	ro.mut.Lock()
	defer ro.mut.Unlock()

	var outputs = make([]omega.Output, 0, len(ro.outputs))
	for _, output := range ro.outputs {
		outputs = append(outputs, output)
	}
	return outputs
}

func (ro *RunningOutput) Stop() {
	ro.mut.Lock()
	defer ro.mut.Unlock()

	close(ro.buffer)
	for _, output := range ro.outputs {
		if running, ok := output.(omega.Running); ok {
			running.Stop()
		}
	}
}
