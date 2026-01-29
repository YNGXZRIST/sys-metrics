package agent

import (
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/stringsparser"
)

var runtimeMetricsTypes = []string{
	common.Alloc, common.BuckHashSys, common.Frees,
	common.GCCPUFraction, common.GCSys,
	common.HeapAlloc, common.HeapIdle,
	common.HeapInuse, common.HeapObjects,
	common.HeapReleased, common.HeapSys,
	common.LastGC, common.Lookups,
	common.MCacheInuse, common.MCacheSys,
	common.MSpanInuse, common.MSpanSys,
	common.Mallocs, common.NextGC,
	common.NumForcedGC, common.NumGC,
	common.OtherSys, common.PauseTotalNs,
	common.StackInuse, common.StackSys,
	common.Sys, common.TotalAlloc,
}
var runtimeMetricsMap = map[string]string{
	strings.ToLower(common.PollCount):     common.PollCount,
	strings.ToLower(common.RandomValue):   common.RandomValue,
	strings.ToLower(common.Alloc):         common.Alloc,
	strings.ToLower(common.BuckHashSys):   common.BuckHashSys,
	strings.ToLower(common.Frees):         common.Frees,
	strings.ToLower(common.GCCPUFraction): common.GCCPUFraction,
	strings.ToLower(common.GCSys):         common.GCSys,
	strings.ToLower(common.HeapAlloc):     common.HeapAlloc,
	strings.ToLower(common.HeapIdle):      common.HeapIdle,
	strings.ToLower(common.HeapInuse):     common.HeapInuse,
	strings.ToLower(common.HeapObjects):   common.HeapObjects,
	strings.ToLower(common.HeapReleased):  common.HeapReleased,
	strings.ToLower(common.HeapSys):       common.HeapSys,
	strings.ToLower(common.LastGC):        common.LastGC,
	strings.ToLower(common.Lookups):       common.Lookups,
	strings.ToLower(common.MCacheInuse):   common.MCacheInuse,
	strings.ToLower(common.MCacheSys):     common.MCacheSys,
	strings.ToLower(common.MSpanInuse):    common.MSpanInuse,
}

type Collector struct {
	Gauges   map[string]*metrics.Gauge
	Counters map[string]*metrics.Counter
	mu       sync.RWMutex
}

func NewCollector() *Collector {
	return &Collector{
		Gauges:   make(map[string]*metrics.Gauge, 27),
		Counters: make(map[string]*metrics.Counter, 2),
	}
}
func (c *Collector) Update() {
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	c.UpdateFromStats(&s)
}

func (c *Collector) UpdateFromStats(s *runtime.MemStats) {
	valueOf := reflect.ValueOf(*s)
	for _, name := range runtimeMetricsTypes {
		v, ok := c.extractFieldValue(valueOf, name)
		if !ok {
			continue
		}
		c.updateOrCreateGauge(name, v)
	}
}
func (c *Collector) updateOrCreateGauge(name string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.Gauges[name]; ok {
		c.Gauges[name].SetValue(value)
		return
	}
	gauge := metrics.NewGauge(name)
	gauge.SetValue(value)
	c.Gauges[name] = gauge
}
func (c *Collector) updateOrCreateCounter(name string, value int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.Counters[name]; ok {
		c.Counters[name].SetValue(value)
		return
	}
	counter := metrics.NewCounter(name)
	counter.SetValue(value)
	c.Counters[name] = counter
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

func (c *Collector) GetGauge(name string) (*metrics.Gauge, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.Gauges[name]
	if !ok || v == nil {
		return metrics.NewGauge(name), false
	}
	return v, true
}

func (c *Collector) SetCounter(name string, value int64) {
	c.updateOrCreateCounter(name, value)
}

func (c *Collector) GetCounter(name string) (*metrics.Counter, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.Counters[name]
	if !ok || v == nil {
		return metrics.NewCounter(name), false
	}
	return v, true
}
func (c *Collector) SetPollCounterMetric() {
	c.updateOrCreateCounter(common.PollCount, 1)
}
func (c *Collector) ResetPollMetric() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.Counters[common.PollCount]; ok && v != nil {
		v.Reset()
	}
}
func (c *Collector) GetPollCountMetric() *metrics.Counter {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.Counters[common.PollCount]; ok && v != nil {
		return v
	}
	return metrics.NewCounter(common.PollCount)
}
func (c *Collector) SetRandomValueMetric() {
	c.updateOrCreateGauge(common.RandomValue, rand.Float64())
}
func GetMetricType(metric string) string {
	lowerMetric := strings.ToLower(metric)
	metricType, ok := runtimeMetricsMap[lowerMetric]
	if !ok {
		metricType = stringsparser.Capitalize(lowerMetric)
	}
	return metricType
}
