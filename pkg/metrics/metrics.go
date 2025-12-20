package metrics

type MetricValue interface {
	int64 | float64
}
type Metric[T MetricValue] interface {
	GetName() string
	GetValue() T
	SetValue(v T)
}
