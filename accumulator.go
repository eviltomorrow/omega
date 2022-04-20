package omega

type Accumulator interface {
	Name() string
	AddMetric([]Metric)
}
