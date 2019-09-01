package log

import (
	"fmt"
	"strings"
	"time"

	"github.com/platinummonkey/isp-monitor/config"
	logger "github.com/platinummonkey/isp-monitor/log"
	"github.com/platinummonkey/isp-monitor/reporters"
	"github.com/platinummonkey/isp-monitor/statistics"
)

func init() {
	reporters.RegisterReporterType("log", NewFromConfig)
}

// Log is a reporter that only logs using the built in logger.
type Log struct {
	name string
}

// NewFromConfig creates a new log reporter from config.
func NewFromConfig(cfg config.Section, _ bool) reporters.Interface {
	return New(cfg.Name)
}

// New creates a new log reporter
func New(name string) *Log {
	if name == "" {
		name = "log"
	}
	return &Log{
		name: name,
	}
}

func (l *Log) tagFormatter(tags []string) string {
	return strings.Join(tags, ", ")
}

// Name returns the name of the reporter
func (l *Log) Name() string {
	return l.name
}

// Timing reports a timing metric
func (l *Log) Timing(metric string, duration time.Duration, tags ...string) {
	logger.Get().Info(fmt.Sprintf("[metric] type=timing name=%s duration=%dms %s", metric, duration/time.Millisecond, l.tagFormatter(tags)))
}

// Count reports a count metric
func (l *Log) Count(metric string, val int64, tags ...string) {
	logger.Get().Info(fmt.Sprintf("[metric] type=count name=%s val=%d %s", metric, val, l.tagFormatter(tags)))
}

// Histogram reports a histogram metric
func (l *Log) Histogram(metric string, val float64, tags ...string) {
	logger.Get().Info(fmt.Sprintf("[metric] type=histogram name=%s val=%f %s", metric, val, l.tagFormatter(tags)))
}

// Gauge reports a gauge metric
func (l *Log) Gauge(metric string, val float64, tags ...string) {
	logger.Get().Info(fmt.Sprintf("[metric] type=gauge name=%s val=%f %s", metric, val, l.tagFormatter(tags)))
}

// Event reports an event
func (l *Log) Event(title string, message string, tags ...string) {
	logger.Get().Info(fmt.Sprintf("[event] title=%s message=%s %s", title, message, l.tagFormatter(tags)))
}

// ReportStatistics implements statistics reporting
func (l *Log) ReportStatistics(stats *statistics.Statistics) {
	for _, stat := range stats.Stats() {
		switch stat.Type() {
		case statistics.EventType:
			l.Event(stat.Event.Title, stat.Event.Message, stat.Event.Tags...)
		case statistics.MetricTypeCount:
			l.Count(stat.Metric.MetricName, stat.Metric.Value.Int(), stat.Metric.Tags...)
		case statistics.MetricTypeGauge:
			l.Gauge(stat.Metric.MetricName, stat.Metric.Value.Float(), stat.Metric.Tags...)
		case statistics.MetricTypeTiming:
			l.Timing(stat.Metric.MetricName, stat.Metric.Value.Duration(), stat.Metric.Tags...)
		case statistics.MetricTypeHistogram:
			l.Histogram(stat.Metric.MetricName, stat.Metric.Value.Float(), stat.Metric.Tags...)
		default:
			// ignore
		}
	}
}
