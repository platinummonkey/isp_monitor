package collectors

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/platinummonkey/isp-monitor/config"
	"github.com/platinummonkey/isp-monitor/log"
	"github.com/platinummonkey/isp-monitor/reporters"
	"github.com/platinummonkey/isp-monitor/statistics"
	"github.com/sparrc/go-ping"
	"go.uber.org/zap"
)

func init() {
	RegisterCollectorType("ping", NewPingerFromConfig)
}

// Pinger is a simple `ping` test
type Pinger struct {
	name            string
	address         string
	count           int
	timeout         time.Duration
	collectInterval time.Duration
	pingInterval    time.Duration
	debug           bool
	packetSize      int
	privileged      bool
}

// PingerOptions are options specific to the pinger
type PingerOptions struct {
	Address    string      `json:"address"`
	Count      json.Number `json:"count"`
	Timeout    string      `json:"timeout"`
	Interval   string      `json:"interval"`
	PacketSize json.Number `json:"packetSize"`
}

// CountInt will return the count as an `int`
func (o PingerOptions) CountInt() int {
	v, err := o.Count.Int64()
	if err != nil {
		return 0
	}
	return int(v)
}

// PacketSizeInt returns the packet size as an `int`
func (o PingerOptions) PacketSizeInt() int {
	v, err := o.PacketSize.Int64()
	if err != nil {
		return 0
	}
	return int(v)
}

// NewPingerFromConfig will create a Pinger from the config Section
func NewPingerFromConfig(cfg config.Section, debug bool) Interface {
	var opts PingerOptions
	if data, err := json.Marshal(cfg.Options); err == nil {
		json.Unmarshal(data, &opts)
	}
	if opts.Address != "" {
		if opts.CountInt() <= 0 {
			opts.Count = "5"
		}
		if opts.PacketSizeInt() <= 0 {
			opts.PacketSize = "0"
		}

		timeout := durationFromString(opts.Timeout, time.Second*5)
		interval := durationFromString(cfg.Interval, time.Second*30)
		pingInterval := durationFromString(opts.Interval, time.Second*30)

		return NewPinger(cfg.Name, opts.Address, opts.CountInt(), timeout, interval, pingInterval, debug, opts.PacketSizeInt())
	}
	return nil
}

// NewPinger will create a new Pinger
func NewPinger(
	name string,
	address string,
	count int,
	timeout time.Duration,
	collectInterval time.Duration,
	pingInterval time.Duration,
	debug bool,
	packetSize int,
) *Pinger {
	return &Pinger{
		name:            name,
		address:         address,
		count:           count,
		collectInterval: collectInterval,
		pingInterval:    pingInterval,
		timeout:         timeout,
		debug:           debug,
		packetSize:      packetSize,
	}
}

// Name returns the name of this Pinger
func (p *Pinger) Name() string {
	return p.name
}

// Collect will collect the ping statistics
func (p *Pinger) Collect() (*statistics.Statistics, error) {
	stats := statistics.NewStatistics()
	pinger, err := ping.NewPinger(p.address)
	if err != nil {
		return stats, err
	}
	pinger.Interval = p.pingInterval
	pinger.Timeout = p.timeout
	pinger.Count = p.count
	pinger.Debug = p.debug
	pinger.Size = p.packetSize
	pinger.SetPrivileged(p.privileged)
	doneChan := make(chan *ping.Statistics, 1)
	pinger.OnFinish = func(statistics *ping.Statistics) {
		doneChan <- statistics
	}
	log.Get().Debug("collecting ping results", zap.String("name", p.name), zap.String("address", p.address))
	go pinger.Run()
	pingStats := <-doneChan
	log.Get().Debug("reporting ping results", zap.String("name", p.name), zap.String("address", p.address))

	tags := []string{
		fmt.Sprintf("address:%s", pingStats.Addr),
		fmt.Sprintf("name:%s", p.name),
	}

	// report RTTs
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeTiming,
				metricPrefix+"pinger.avg_rtt",
				statistics.NewDurationValue(pingStats.AvgRtt),
				tags...,
			),
			nil,
		),
	)
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeTiming,
				metricPrefix+"pinger.max_rtt",
				statistics.NewDurationValue(pingStats.MaxRtt),
				tags...,
			),
			nil,
		),
	)
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeTiming,
				metricPrefix+"pinger.min_rtt",
				statistics.NewDurationValue(pingStats.MinRtt),
				tags...,
			),
			nil,
		),
	)
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeTiming,
				metricPrefix+"pinger.stddev_rtt",
				statistics.NewDurationValue(pingStats.StdDevRtt),
				tags...,
			),
			nil,
		),
	)

	// packet info
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeCount,
				metricPrefix+"pinger.packets_sent",
				statistics.NewIntValue(int64(pingStats.PacketsSent)),
				tags...,
			),
			nil,
		),
	)
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeCount,
				metricPrefix+"pinger.packets_recv",
				statistics.NewIntValue(int64(pingStats.PacketsRecv)),
				tags...,
			),
			nil,
		),
	)
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeHistogram,
				metricPrefix+"pinger.packet_loss",
				statistics.NewFloatValue(pingStats.PacketLoss),
				tags...,
			),
			nil,
		),
	)

	return stats, nil
}

// Run will run the Pinger routine in the background
func (p *Pinger) Run(reporters map[string]reporters.Interface) {
	tags := []string{
		fmt.Sprintf("address:%s", p.address),
		fmt.Sprintf("name:%s", p.name),
	}

	collect := func() {
		stats, err := p.Collect()
		if err != nil {
			log.Get().Warn("failed to execute ping", zap.String("name", p.name), zap.String("address", p.address), zap.Error(err))
			// error statistic
			stats.Add(
				statistics.NewStatistic(
					statistics.NewMetric(
						statistics.MetricTypeCount,
						metricPrefix+"pinger."+collectFailureSuffix,
						statistics.NewIntValue(1),
						tags...,
					),
					nil,
				),
			)
		}
		for _, reporter := range reporters {
			reporter.ReportStatistics(stats)
		}
	}

	go func() {
		ticker := time.NewTicker(p.collectInterval)
		collect()
		for {
			select {
			case <-ticker.C:
				collect()
			}
		}
	}()
}
