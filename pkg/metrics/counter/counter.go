package counter

type Metric struct {
	name  string
	value int64
}

func NewCounter(n string) *Metric {
	return &Metric{name: n}
}
func (c *Metric) GetName() string {

	return c.name
}

func (c *Metric) GetValue() int64 {
	return c.value
}

func (c *Metric) SetValue(v int64) {
	c.value += v
}
