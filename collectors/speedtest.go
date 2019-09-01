package collectors

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/platinummonkey/isp-monitor/config"
	"github.com/platinummonkey/isp-monitor/log"
	"github.com/platinummonkey/isp-monitor/reporters"
	"github.com/platinummonkey/isp-monitor/statistics"
	"github.com/surol/speedtest-cli/speedtest"
	"go.uber.org/zap"
)

func init() {
	RegisterCollectorType("speedtest", NewSpeedTestFromConfig)
}

// SpeedTest is the speedtest collector
type SpeedTest struct {
	client   speedtest.Client
	interval time.Duration
}

// SpeedTestOptions are options specific to SpeedTest
type SpeedTestOptions struct {
	Secure  bool   `json:"secure"`
	Timeout string `json:"timeout"`
}

// NewSpeedTestFromConfig will create a new SpeedTest from config
func NewSpeedTestFromConfig(cfg config.Section, debug bool) Interface {
	var opts SpeedTestOptions
	if data, err := json.Marshal(cfg.Options); err == nil {
		json.Unmarshal(data, &opts)
	}

	timeout := durationFromString(opts.Timeout, time.Second*5)
	interval := durationFromString(cfg.Interval, time.Second*30)

	return NewSpeedTest(opts.Secure, timeout, interval, debug)
}

// NewSpeedTest will create a new SpeedTest
func NewSpeedTest(
	secure bool,
	timeout time.Duration,
	interval time.Duration,
	debug bool,
) *SpeedTest {
	opts := &speedtest.Opts{
		SpeedInBytes: false,
		Quiet:        !debug,
		List:         false,
		Server:       0,
		Interface:    "",
		Timeout:      timeout,
		Secure:       secure,
		Help:         false,
		Version:      false,
	}
	client := speedtest.NewClient(opts)
	return &SpeedTest{
		client:   client,
		interval: interval,
	}
}

// Name returns the name of this test.
func (c *SpeedTest) Name() string {
	return "speedtest"
}

// Collect will run the test and report statistics.
func (c *SpeedTest) Collect() (*statistics.Statistics, error) {
	log.Get().Debug("collecting speedtest results")
	stats := statistics.NewStatistics()
	config, err := c.client.Config()
	if err != nil {
		return stats, err
	}
	log.Get().Debug(fmt.Sprintf("speedtest - Testing from %s (%s)...\n", config.Client.ISP, config.Client.IP))
	servers, err := c.client.ClosestServers()
	if err != nil {
		return stats, fmt.Errorf("Failed to load server list: %v\n", err)
	}
	server := servers.MeasureLatencies(
		speedtest.DefaultLatencyMeasureTimes,
		speedtest.DefaultErrorLatency).First()
	log.Get().Debug(
		fmt.Sprintf("speedtest - Hosted by %s (%s) [%.2f km]: %d ms",
			server.Sponsor,
			server.Name,
			server.Distance,
			server.Latency/time.Millisecond,
		),
	)
	log.Get().Debug("reporting speedtest results")

	tags := []string{
		fmt.Sprintf("server_id:%d", server.ID),
		fmt.Sprintf("server_sponsor:%s", server.Sponsor),
	}

	// report distance
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeGauge,
				metricPrefix+"speedtest.server_distance",
				statistics.NewFloatValue(server.Distance),
				tags...,
			),
			nil,
		),
	)
	// report latency
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeHistogram,
				metricPrefix+"speedtest.server_latency",
				statistics.NewDurationValue(server.Latency/time.Millisecond),
				tags...,
			),
			nil,
		),
	)

	// report download speed
	downloadSpeed := float64(server.DownloadSpeed()) / (1 << 17)
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeTiming,
				metricPrefix+"speedtest.download_speed",
				statistics.NewFloatValue(downloadSpeed), tags...),
			nil,
		),
	)

	// report upload speed
	uploadSpeed := float64(server.UploadSpeed()) / (1 << 17)
	stats.Add(
		statistics.NewStatistic(
			statistics.NewMetric(
				statistics.MetricTypeTiming,
				metricPrefix+"speedtest.upload_speed",
				statistics.NewFloatValue(uploadSpeed), tags...),
			nil,
		),
	)

	return stats, nil
}

// Run will run the test in the background.
func (c *SpeedTest) Run(reporters map[string]reporters.Interface) {
	collect := func() {
		stats, err := c.Collect()
		if err != nil {
			log.Get().Warn("failed to execute speedtest", zap.Error(err))
			// error statistic
			stats.Add(
				statistics.NewStatistic(
					statistics.NewMetric(
						statistics.MetricTypeCount,
						metricPrefix+"speedtest."+collectFailureSuffix,
						statistics.NewIntValue(1),
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
		ticker := time.NewTicker(c.interval)
		collect()
		for {
			select {
			case <-ticker.C:
				collect()
			}
		}
	}()
}
