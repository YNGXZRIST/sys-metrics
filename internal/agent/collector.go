package agent

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sys-metrics/internal/common"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/stringsparser"
	"sys-metrics/pkg/workerpool"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
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
	strings.ToLower(common.TotalMemory):   common.TotalMemory,
	strings.ToLower(common.FreeMemory):    common.FreeMemory,
}

// generate:reset

// Collector stores gauges and counters in memory using pools for parallel OS and runtime sampling.
type Collector struct {
	Gauges        map[string]*metrics.Gauge
	Counters      map[string]*metrics.Counter
	collectorPool *workerpool.Pool
	updaterPool   *workerpool.Pool
	ctx           context.Context
	mu            sync.Mutex
}

// NewCollector initializes metric maps and worker pools for CPU/memory and runtime collection.
func NewCollector(ctx context.Context, rateLimit int) *Collector {
	collectorPool := workerpool.NewPool(rateLimit)
	collectorPool.StartBg(ctx)
	updaterPool := workerpool.NewPool(rateLimit)
	updaterPool.StartBg(ctx)
	return &Collector{
		Gauges:        make(map[string]*metrics.Gauge, 27),
		Counters:      make(map[string]*metrics.Counter, 2),
		collectorPool: collectorPool,
		updaterPool:   updaterPool,
		ctx:           ctx,
	}
}

// Update enqueues asynchronous memory, CPU, and runtime sampling without waiting for results.
func (c *Collector) Update() error {
	sysTask := c.getSysTask()
	memTask := c.getMemTask()
	sysTask.NeedResult = false
	memTask.NeedResult = false
	c.collectorPool.Add(sysTask)
	c.collectorPool.Add(memTask)
	return nil
}

// UpdateSync performs the same sampling as Update but waits for tasks to finish.
func (c *Collector) UpdateSync() error {
	sysTask := c.getSysTask()
	memTask := c.getMemTask()
	sysTask.NeedResult = true
	memTask.NeedResult = true
	c.collectorPool.Add(sysTask)
	c.collectorPool.Add(memTask)
	sysRes := c.collectorPool.Get()
	if sysRes.Err != nil {
		return sysRes.Err
	}
	memRes := c.collectorPool.Get()
	if memRes.Err != nil {
		return memRes.Err
	}
	return nil

}

func (c *Collector) getMemTask() *workerpool.Task {
	memTask := workerpool.NewTask(func(a any) (any, error) {
		v, err := mem.VirtualMemory()
		if err != nil {
			err = labelerrors.NewLabelError("COLLECT", fmt.Errorf("error getting mem.VirtualMemory: %w", err))
			return nil, err
		}
		m := make(map[string]float64)
		m[common.TotalMemory] = float64(v.Total)
		m[common.FreeMemory] = float64(v.Free)
		cpuPercents, _ := cpu.Percent(0, true)
		for i, p := range cpuPercents {
			name := fmt.Sprintf("CPUutilization%d", i+1)
			m[name] = p
		}
		c.UpdateFromStats(m)
		return m, nil
	})
	return memTask
}
func (c *Collector) getSysTask() *workerpool.Task {
	sysTask := workerpool.NewTask(func(a any) (any, error) {
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		valueOf := reflect.ValueOf(s)
		m := make(map[string]float64)
		for _, name := range runtimeMetricsTypes {
			v, ok := c.extractFieldValue(valueOf, name)
			if !ok {
				continue
			}
			m[name] = v
		}
		c.UpdateFromStats(m)
		return m, nil
	})
	return sysTask
}

func (c *Collector) UpdateFromStats(maps map[string]float64) {
	for n, v := range maps {
		c.updateOrCreateGauge(n, v)
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
	c.mu.Lock()
	defer c.mu.Unlock()
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
	c.mu.Lock()
	defer c.mu.Unlock()
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
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.Counters[common.PollCount]; ok && v != nil {
		return v
	}
	return metrics.NewCounter(common.PollCount)
}
func (c *Collector) SetRandomValueMetric() {
	c.updateOrCreateGauge(common.RandomValue, rand.Float64())
}

// GetMetricType normalizes a metric name: known runtime names via map, otherwise Capitalize.
func GetMetricType(metric string) string {
	lowerMetric := strings.ToLower(metric)
	metricType, ok := runtimeMetricsMap[lowerMetric]
	if !ok {
		metricType = stringsparser.Capitalize(lowerMetric)
	}
	return metricType
}
