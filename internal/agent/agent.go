package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/eviltomorrow/omega"
	"github.com/eviltomorrow/omega/internal/conf"
	"github.com/eviltomorrow/omega/internal/output"
	"github.com/eviltomorrow/omega/pkg/plugins"
	"github.com/eviltomorrow/omega/pkg/ticker"
	"github.com/eviltomorrow/omega/pkg/zlog"
	"go.uber.org/zap"

	_ "github.com/eviltomorrow/omega/pkg/plugins/cpu"
	_ "github.com/eviltomorrow/omega/pkg/plugins/disk"
	_ "github.com/eviltomorrow/omega/pkg/plugins/diskio"
	_ "github.com/eviltomorrow/omega/pkg/plugins/host"
	_ "github.com/eviltomorrow/omega/pkg/plugins/mem"
	_ "github.com/eviltomorrow/omega/pkg/plugins/net"
	_ "github.com/eviltomorrow/omega/pkg/plugins/processes"
	_ "github.com/eviltomorrow/omega/pkg/plugins/swap"
)

type Agent struct {
	config *conf.Config
}

func NewAgent(config *conf.Config) (*Agent, error) {
	return &Agent{config: config}, nil
}

func (a *Agent) Run(ctx context.Context, signal chan struct{}) error {
	if err := ConfigPlugins(a.config); err != nil {
		return fmt.Errorf("config plugins failure, nest error: %v", err)
	}

	var wg sync.WaitGroup

	ou, err := output.NewGrpcClient(a.config.Global.GroupName, a.config.Global.EtcdEndpoints)
	if err != nil {
		return err
	}
	defer ou.Close()

	clearFunc, err := RunningOutputPool.RegisterOutput(ou)
	if err != nil {
		return err
	}
	defer clearFunc()

	if err := RunningOutputPool.Start(); err != nil {
		return fmt.Errorf("start output failure, nest error: %v", err)
	}
	for name, plugin := range plugins.Repository {
		var (
			ac     = NewAccumulator(name, RunningOutputPool.Buffer())
			ticker = ticker.NewAlignedTicker(time.Now(), a.config.Agent.Period.Duration, 0, 0)
		)

		wg.Add(1)
		go func(ticker omega.Ticker, ac omega.Accumulator, plugin plugins.Collector) {
			gatherLoop(ticker, ac, plugin)
			wg.Done()
		}(ticker, ac, plugin)
		registerClearFunc(ticker.Stop)
	}

	signal <- struct{}{}

	for range ctx.Done() {
		runClearFuncs()
	}
	wg.Wait()
	resetClearFuncRepo()

	return nil
}

func ConfigPlugins(config *conf.Config) error {
	for name, plugin := range plugins.Repository {
		cp, ok := config.Plugins[name]
		if !ok {
			continue
		}
		if err := json.Unmarshal(cp.Byte(), plugin); err != nil {
			return err
		}
		plugins.Register(name, plugin)
	}
	return nil
}

func gatherLoop(ticker omega.Ticker, ac omega.Accumulator, plugin plugins.Collector) {
	for range ticker.Elapsed() {
		metrics, err := plugin.Gather()
		if err != nil {
			zlog.Error("Gather metric failure", zap.String("name", ac.Name()), zap.Error(err))
		} else {
			ac.AddMetric(metrics)
		}
	}
}

var clearFuncRepo []func() = make([]func(), 0, 8)

func registerClearFunc(f func()) {
	clearFuncRepo = append(clearFuncRepo, f)
}

func runClearFuncs() {
	for _, f := range clearFuncRepo {
		f()
	}
}

func resetClearFuncRepo() {
	clearFuncRepo = clearFuncRepo[:0]
}
