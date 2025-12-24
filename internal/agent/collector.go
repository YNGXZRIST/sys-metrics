package agent

import (
	"math/rand"
	"reflect"
	"runtime"
)

var runtimeMetricsTypes = []string{
	"Alloc", "BuckHashSys", "Frees",
	"GCCPUFraction", "GCSys",
	"HeapAlloc", "HeapIdle",
	"HeapInuse", "HeapObjects",
	"HeapReleased", "HeapSys",
	"LastGC", "Lookups",
	"MCacheInuse", "MCacheSys",
	"MSpanInuse", "MSpanSys",
	"Mallocs", "NextGC",
	"NumForcedGC", "NumGC",
	"OtherSys", "PauseTotalNs",
	"StackInuse", "StackSys",
	"Sys", "TotalAlloc",
}

type Collector struct {
	metrics map[string]map[string]float64
}

func NewCollector() *Collector {
	m := make(map[string]map[string]float64)
	m["gauge"] = make(map[string]float64, 27)
	m["counter"] = make(map[string]float64, 2)
	return &Collector{
		metrics: m,
	}
}
func (c *Collector) Update() {
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	v := reflect.ValueOf(s)
	for _, n := range runtimeMetricsTypes {
		f := v.FieldByName(n)
		switch f.Kind() {
		case reflect.Uint64, reflect.Uint32:
			c.metrics["gauge"][n] = float64(f.Uint())
		case reflect.Float64:
			c.metrics["gauge"][n] = f.Float()
		default:
			continue
		}
	}
	return
}
func (c *Collector) SetPoolMetric() {
	c.metrics["counter"]["PollCount"]++
}
func (c *Collector) ResetPoolMetric() {
	c.metrics["counter"]["PollCount"] = 0
}
func (c *Collector) SetRandomValueMetric() {
	c.metrics["gauge"]["RandomValue"] = rand.Float64()
}
