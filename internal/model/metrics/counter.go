package metrics

import "sys-metrics/internal/common"

// Counter is a counter metric; delta accumulates via SetValue.
type Counter struct {
	Metrics
}

// SetValue adds v to the delta (or sets it on first call) and returns the new value.
func (c *Counter) SetValue(v int64) int64 {
	if c.Delta == nil {
		c.Delta = &v
		return *c.Delta
	}
	*c.Delta += v
	return *c.Delta
}

// Reset clears the counter delta.
func (c *Counter) Reset() {
	if c.Delta == nil {
		c.Delta = new(int64(0))

	}
	*c.Delta = 0
}

// NewCounter creates a counter with zero delta and type common.Counter.
func NewCounter(name string) *Counter {
	return &Counter{Metrics{ID: name, MType: common.Counter, Delta: new(int64(0))}}
}
