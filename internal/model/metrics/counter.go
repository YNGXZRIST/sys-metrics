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

func NewCounter(name string) *Counter {
	delta := int64(0)
	return &Counter{Metrics{ID: name, MType: common.Counter, Delta: &delta}}
}
