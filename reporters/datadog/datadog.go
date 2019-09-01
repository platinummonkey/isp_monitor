package datadog

import (
	"encoding/json"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/platinummonkey/isp-monitor/config"
	"github.com/platinummonkey/isp-monitor/reporters"
	"github.com/platinummonkey/isp-monitor/statistics"
)

// DataDog implements a dogstatsd reporter interface
type DataDog struct {
	name   string
	client *statsd.Client
}

// DataDogOptions are the options specific to the DataDog reporter
type DataDogOptions struct {
	Address               string   `json:"address"`
	Namespace             string   `json:"namespace"`
	Tags                  []string `json:"tags"`
	Buffered              bool     `json:"buffered"`
	MaxMessagesPerPayload int      `json:"max_messages_per_payload"`
	AsyncUDS              bool     `json:"async_uds"`
	WriteTimeoutUDS       string   `json:"write_timeout_uds"`
}

// NewFromConfig will return a new DataDog reporter from the provided config.
func NewFromConfig(cfg config.Section, debug bool) reporters.Interface {
	var opts DataDogOptions
	if data, err := json.Marshal(cfg.Options); err == nil {
		json.Unmarshal(data, &opts)
	}

	statsOpts := make([]statsd.Option, 0)

	writeTimeoutUDS := time.Duration(0)
	if opts.WriteTimeoutUDS != "" {
		dur, err := time.ParseDuration(opts.WriteTimeoutUDS)
		if err == nil && dur >= 0 {
			writeTimeoutUDS = dur
		}
	}

	if writeTimeoutUDS > 0 {
		statsOpts = append(statsOpts, statsd.WithWriteTimeoutUDS(writeTimeoutUDS))
	}
	if opts.Namespace != "" {
		statsOpts = append(statsOpts, statsd.WithNamespace(opts.Namespace))
	}
	if opts.AsyncUDS {
		statsOpts = append(statsOpts, statsd.WithAsyncUDS())
	}
	if opts.Buffered {
		statsOpts = append(statsOpts, statsd.Buffered())
	}
	if opts.MaxMessagesPerPayload > 0 {
		statsOpts = append(statsOpts, statsd.WithMaxMessagesPerPayload(opts.MaxMessagesPerPayload))
	}
	if opts.Tags != nil && len(opts.Tags) > 0 {
		statsOpts = append(statsOpts, statsd.WithTags(opts.Tags))
	}

	client, err := statsd.New(opts.Address, statsOpts...)
	if err != nil {
		return nil
	}
	return New(cfg.Name, client)
}

// New returns a DataDog reporter using the provided client.
func New(name string, client *statsd.Client) *DataDog {
	if name == "" {
		name = "datadog"
	}
	return &DataDog{
		name:   name,
		client: client,
	}
}

// Name the name of the reporter
func (d *DataDog) Name() string {
	return d.name
}

// Timing reports a timing metric
func (d *DataDog) Timing(metric string, duration time.Duration, tags ...string) {
	d.client.Timing(metric, duration, tags, 1.0)
}

// Histogram reports a histogram metric
func (d *DataDog) Histogram(metric string, value float64, tags ...string) {
	d.client.Histogram(metric, value, tags, 1.0)
}

// Count reports a count metric
func (d *DataDog) Count(metric string, value int64, tags ...string) {
	d.client.Count(metric, value, tags, 1.0)
}

// Gauge reports a gauge metric
func (d *DataDog) Gauge(metric string, value float64, tags ...string) {
	d.client.Gauge(metric, value, tags, 1.0)
}

// Event reports an event
func (d *DataDog) Event(title string, message string, tags ...string) {
	e := statsd.NewEvent(title, message)
	e.Tags = tags
	d.client.Event(e)
}

// ReportStatistics implements statistics reporting
func (d *DataDog) ReportStatistics(stats *statistics.Statistics) {
	for _, stat := range stats.Stats() {
		switch stat.Type() {
		case statistics.EventType:
			d.Event(stat.Event.Title, stat.Event.Message, stat.Event.Tags...)
		case statistics.MetricTypeCount:
			d.Count(stat.Metric.MetricName, stat.Metric.Value.Int(), stat.Metric.Tags...)
		case statistics.MetricTypeGauge:
			d.Gauge(stat.Metric.MetricName, stat.Metric.Value.Float(), stat.Metric.Tags...)
		case statistics.MetricTypeTiming:
			d.Timing(stat.Metric.MetricName, stat.Metric.Value.Duration(), stat.Metric.Tags...)
		case statistics.MetricTypeHistogram:
			d.Histogram(stat.Metric.MetricName, stat.Metric.Value.Float(), stat.Metric.Tags...)
		default:
			// ignore
		}
	}
}
