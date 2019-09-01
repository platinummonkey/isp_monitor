package reporters

import (
	"sync"
	"time"

	"github.com/platinummonkey/isp-monitor/config"
	"github.com/platinummonkey/isp-monitor/statistics"
)

// Interface is the reporter interface
type Interface interface {
	Timing(metric string, duration time.Duration, tags ...string)
	Count(metric string, val int64, tags ...string)
	Histogram(metric string, val float64, tags ...string)
	Gauge(metric string, val float64, tags ...string)
	Event(title string, message string, tags ...string)
	ReportStatistics(statistics *statistics.Statistics)

	Name() string
}

var registeredReporters map[string]func(config.Section, bool) Interface
var mu sync.RWMutex

// RegisterReporterType will register a reporter type
// this makes this extensible to outside projects.
func RegisterReporterType(collectorType string, registerFunc func(config.Section, bool) Interface) {
	mu.Lock()
	registeredReporters[collectorType] = registerFunc
	mu.Unlock()
}

// CreateReporterFromConfig will create a reporter from the given config
func CreateReporterFromConfig(cfg config.Section, debug bool) Interface {
	mu.RLock()
	defer mu.RUnlock()
	if creatorFunc, ok := registeredReporters[cfg.Type]; ok {
		return creatorFunc(cfg, debug)
	}
	return nil
}
