package agent

import (
	"math/rand"
	"reflect"
	"runtime"
	"sys-metrics/internal/common"
)

const (
	PollCount   = "PollCount"
	RandomValue = "RandomValue"

	Alloc         = "Alloc"
	BuckHashSys   = "BuckHashSys"
	Frees         = "Frees"
	GCCPUFraction = "GCCPUFraction"
	GCSys         = "GCSys"
	HeapAlloc     = "HeapAlloc"
	HeapIdle      = "HeapIdle"
	HeapInuse     = "HeapInuse"
	HeapObjects   = "HeapObjects"
	HeapReleased  = "HeapReleased"
	HeapSys       = "HeapSys"
	LastGC        = "LastGC"
	Lookups       = "Lookups"
	MCacheInuse   = "MCacheInuse"
	MCacheSys     = "MCacheSys"
	MSpanInuse    = "MSpanInuse"
	MSpanSys      = "MSpanSys"
	Mallocs       = "Mallocs"
	NextGC        = "NextGC"
	NumForcedGC   = "NumForcedGC"
	NumGC         = "NumGC"
	OtherSys      = "OtherSys"
	PauseTotalNs  = "PauseTotalNs"
	StackInuse    = "StackInuse"
	StackSys      = "StackSys"
	Sys           = "Sys"
	TotalAlloc    = "TotalAlloc"
)

var runtimeMetricsTypes = []string{
	Alloc, BuckHashSys, Frees,
	GCCPUFraction, GCSys,
	HeapAlloc, HeapIdle,
	HeapInuse, HeapObjects,
	HeapReleased, HeapSys,
	LastGC, Lookups,
	MCacheInuse, MCacheSys,
	MSpanInuse, MSpanSys,
	Mallocs, NextGC,
	NumForcedGC, NumGC,
	OtherSys, PauseTotalNs,
	StackInuse, StackSys,
	Sys, TotalAlloc,
}

type Collector struct {
	metrics map[string]map[string]float64
}

func NewCollector() *Collector {
	m := make(map[string]map[string]float64)
	m[common.Gauge] = make(map[string]float64, 27)
	m[common.Counter] = make(map[string]float64, 2)
	return &Collector{
		metrics: m,
	}
}
func (c *Collector) Update() {
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	c.UpdateFromStats(&s)
}

func (c *Collector) UpdateFromStats(s *runtime.MemStats) {
	v := reflect.ValueOf(*s)
	for _, name := range runtimeMetricsTypes {
		if value, ok := c.extractFieldValue(v, name); ok {
			c.SetGauge(name, value)
		}
	}
}

func (c *Collector) extractFieldValue(v reflect.Value, name string) (float64, bool) {
	f := v.FieldByName(name)
	switch f.Kind() {
	case reflect.Uint64, reflect.Uint32:
		return float64(f.Uint()), true
	case reflect.Float64:
		return f.Float(), true
	default:
		return 0, false
	}
}

func (c *Collector) SetGauge(name string, value float64) {
	c.metrics[common.Gauge][name] = value
}

func (c *Collector) GetGauge(name string) (float64, bool) {
	v, ok := c.metrics[common.Gauge][name]
	return v, ok
}

func (c *Collector) SetCounter(name string, value float64) {
	c.metrics[common.Counter][name] = value
}

func (c *Collector) GetCounter(name string) (float64, bool) {
	v, ok := c.metrics[common.Counter][name]
	return v, ok
}
func (c *Collector) SetPollCounterMetric() {
	c.metrics[common.Counter][PollCount]++
}
func (c *Collector) ResetPollMetric() {
	c.metrics[common.Counter][PollCount] = 0
}
func (c *Collector) GetPollCountMetric() float64 {
	return c.metrics[common.Counter][PollCount]
}
func (c *Collector) SetRandomValueMetric() {
	c.metrics[common.Gauge][RandomValue] = rand.Float64()
}
func (c *Collector) GetRandomValueMetric() float64 {
	return c.metrics[common.Gauge][RandomValue]
}
