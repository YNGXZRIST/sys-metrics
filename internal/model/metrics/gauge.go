package metrics

import "sys-metrics/internal/common"

// Gauge is a gauge metric embedding Metrics for ID and serialization.
type Gauge struct {
	Metrics
}

// SetValue sets the current gauge value and returns it.
func (g *Gauge) SetValue(v float64) float64 {
	if g.Value == nil {
		g.Value = new(float64)
	}
	*g.Value = v
	return *g.Value
}

// NewGauge creates a gauge with the given name and type common.Gauge.
func NewGauge(name string) *Gauge {
	return &Gauge{Metrics{ID: name, MType: common.Gauge}}
}
