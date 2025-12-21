package metrics

type Gauge struct {
	Metrics
}

func (g *Gauge) SetValue(v float64) float64 {
	g.Value = &v
	return *g.Value
}

func NewGauge(name string) *Gauge {
	return &Gauge{Metrics{ID: name, MType: MTypeGauge}}
}
