package collectors

import (
	"sync"
	"time"

	"github.com/platinummonkey/isp-monitor/config"
	"github.com/platinummonkey/isp-monitor/reporters"
	"github.com/platinummonkey/isp-monitor/statistics"
)

const metricPrefix = "isp_monitor."
const collectFailureSuffix = "collect_failure"

// Interface defines a collector interface.
type Interface interface {
	Run(reporters map[string]reporters.Interface)
	Collect() (*statistics.Statistics, error)
	Name() string
}

var registeredCollectors map[string]func(config.Section, bool) Interface
var mu sync.RWMutex

// RegisterCollectorType will register a collector type
// this makes this extensible to outside projects.
func RegisterCollectorType(collectorType string, registerFunc func(config.Section, bool) Interface) {
	mu.Lock()
	registeredCollectors[collectorType] = registerFunc
	mu.Unlock()
}

// CreateCollectorFromConfig will create a collector from the given config
func CreateCollectorFromConfig(cfg config.Section, debug bool) Interface {
	mu.RLock()
	defer mu.RUnlock()
	if creatorFunc, ok := registeredCollectors[cfg.Type]; ok {
		return creatorFunc(cfg, debug)
	}
	return nil
}

func durationFromString(s string, defaultDuration time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err == nil && d > 0 {
		return d
	}
	return defaultDuration
}
