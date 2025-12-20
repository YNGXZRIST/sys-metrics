package gauge

type Metric struct {
	Name  string
	Value float64
}

func NewGauge(n string) *Metric {
	return &Metric{Name: n}
}
func (c *Metric) GetName() string {
	return c.Name
}
func (c *Metric) GetValue() float64 {
	return c.Value
}
func (c *Metric) SetValue(v float64) {
	c.Value = v
}
