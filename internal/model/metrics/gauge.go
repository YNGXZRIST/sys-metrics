package metrics

import "sys-metrics/internal/common"

type Gauge struct {
	Metrics
}

func (g *Gauge) SetValue(v float64) float64 {
	if g.Value == nil {
		g.Value = new(float64)
	}
	*g.Value = v
	return *g.Value
}

func NewGauge(name string) *Gauge {
	return &Gauge{Metrics{ID: name, MType: common.Gauge}}
}
