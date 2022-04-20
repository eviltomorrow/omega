package omega

import "time"

type ValueType int

const (
	_ ValueType = iota
	Counter
	Gauge
	StateSet
	Info
	Summary
	Histogram
	GaugeHistogram
	Unknown
)

type Tag struct {
	Key   string
	Value string
}

type Field struct {
	Key   string
	Value interface{}
}

type Metric interface {
	Name() string

	SetName(name string)

	Tags() map[string]string

	AddTag(key, value string)

	Fields() map[string]interface{}

	AddField(key string, value interface{})

	Time() time.Time

	SetTime(t time.Time)

	Type() ValueType

	String() string
}
