package statistics

import (
	"fmt"
	"sync"
	"time"
)

// Type is the type of the statistics data
type Type string

// Supported types of statistics data
const (
	EventType           Type = "event"
	MetricTypeGauge     Type = "gauge"
	MetricTypeCount     Type = "count"
	MetricTypeHistogram Type = "histogram"
	MetricTypeTiming    Type = "timing"
)

// Value is a generic value holder
type Value struct {
	stringValue *string
	floatValue  *float64
	intValue    *int64
	uintValue   *uint64
	duration    *time.Duration
}

// NewStringValue holds a string value
func NewStringValue(val string) Value {
	return Value{
		stringValue: &val,
	}
}

// NewIntValue holds an int value
func NewIntValue(val int64) Value {
	return Value{
		intValue: &val,
	}
}

// NewUintValue holds an uint value
func NewUintValue(val uint64) Value {
	return Value{
		uintValue: &val,
	}
}

// NewFloatValue holds a float value
func NewFloatValue(val float64) Value {
	return Value{
		floatValue: &val,
	}
}

// NewDurationValue holds a duration value
func NewDurationValue(val time.Duration) Value {
	return Value{
		duration: &val,
	}
}

// Duration returns the value as a `time.Duration`
func (v Value) Duration() time.Duration {
	if v.duration != nil {
		return *v.duration
	}
	if v.floatValue != nil {
		return time.Duration(*v.floatValue)
	}
	if v.intValue != nil {
		return time.Duration(*v.intValue)
	}
	if v.uintValue != nil {
		return time.Duration(*v.uintValue)
	}
	return 0
}

// String returns the value as a `string`
func (v Value) String() string {
	if v.stringValue != nil {
		return *v.stringValue
	}
	if v.floatValue != nil {
		return fmt.Sprintf("%f", *v.floatValue)
	}
	if v.intValue != nil {
		return fmt.Sprintf("%d", *v.intValue)
	}
	if v.uintValue != nil {
		return fmt.Sprintf("%d", *v.uintValue)
	}
	if v.duration != nil {
		return fmt.Sprintf("%d ms", (*v.duration)/time.Millisecond)
	}
	return ""
}

// Float returns the value as a `float64`
func (v Value) Float() float64 {
	if v.floatValue != nil {
		return *v.floatValue
	}
	if v.intValue != nil {
		return float64(*v.intValue)
	}
	if v.uintValue != nil {
		return float64(*v.uintValue)
	}
	if v.duration != nil {
		return float64(*v.duration)
	}
	return 0
}

// Int returns the value as an `int64`
func (v Value) Int() int64 {
	if v.intValue != nil {
		return *v.intValue
	}
	if v.uintValue != nil {
		return int64(*v.uintValue)
	}
	if v.floatValue != nil {
		return int64(*v.floatValue)
	}
	if v.duration != nil {
		return int64(*v.duration)
	}
	return 0
}

// Metric is a metric statistic
type Metric struct {
	MetricType Type
	MetricName string
	Value      Value
	Tags       []string
}

// NewMetric creates a new metric
func NewMetric(metricType Type, metricName string, value Value, tags ...string) *Metric {
	return &Metric{
		MetricType: metricType,
		MetricName: metricName,
		Value:      value,
		Tags:       tags,
	}
}

// Event is an event statistic
type Event struct {
	Title   string
	Message string
	Tags    []string
}

// NewEvent creates a new event
func NewEvent(title string, message string, tags ...string) *Event {
	return &Event{
		Title:   title,
		Message: message,
		Tags:    tags,
	}
}

// Statistic a metric collection
type Statistic struct {
	Metric *Metric
	Event  *Event
}

// Type returns the statistic type
func (s *Statistic) Type() Type {
	if s.Event != nil {
		return EventType
	}
	if s.Metric != nil {
		return s.Metric.MetricType
	}
	return "unknown"
}

// NewStatistic creates a new statistic
func NewStatistic(metric *Metric, event *Event) *Statistic {
	return &Statistic{
		Metric: metric,
		Event:  event,
	}
}

// Statistics contains many statistics. Thread-safe.
type Statistics struct {
	statistics []*Statistic
	mu         sync.RWMutex
}

// NewStatistics creates a new Statistics bucket.
func NewStatistics() *Statistics {
	return &Statistics{
		statistics: make([]*Statistic, 0),
	}
}

// Add appends to the statistics
func (s *Statistics) Add(stat *Statistic) {
	s.mu.Lock()
	s.statistics = append(s.statistics, stat)
	s.mu.Unlock()
}

// Stats returns the current bucket of statistics.
func (s *Statistics) Stats() []*Statistic {
	s.mu.RLock()
	stats := append([]*Statistic{}, s.statistics...)
	s.mu.RUnlock()
	return stats
}
