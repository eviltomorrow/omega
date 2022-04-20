package omega

type Output interface {
	Connect() error
	Close() error
	WriteMetric(metric []Metric) error
}
