package metrics

import "sys-metrics/internal/common"

type Counter struct {
	Metrics
}

func (c *Counter) SetValue(v int64) int64 {
	if c.Delta == nil {
		c.Delta = &v
		return *c.Delta
	}
	*c.Delta += v
	return *c.Delta
}
func (c *Counter) Reset() {
	if c.Delta == nil {
		c.Delta = new(int64(0))

	}
	*c.Delta = 0
}
func NewCounter(name string) *Counter {
	return &Counter{Metrics{ID: name, MType: common.Counter, Delta: new(int64(0))}}
}
