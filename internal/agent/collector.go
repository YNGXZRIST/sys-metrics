package agent

import (
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"sys-metrics/internal/common"
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
	c.metrics[common.Counter][common.PollCount]++
}
func (c *Collector) ResetPollMetric() {
	c.metrics[common.Counter][common.PollCount] = 0
}
func (c *Collector) GetPollCountMetric() float64 {
	return c.metrics[common.Counter][common.PollCount]
}
func (c *Collector) SetRandomValueMetric() {
	c.metrics[common.Gauge][common.RandomValue] = rand.Float64()
}
func (c *Collector) GetRandomValueMetric() float64 {
	return c.metrics[common.Gauge][common.RandomValue]
}
func GetMetricType(metric string) string {
	lowerMetric := strings.ToLower(metric)
	metricType, ok := runtimeMetricsMap[lowerMetric]
	if !ok {
		metricType = stringsparser.Capitalize(lowerMetric)
	}
	return metricType
}
