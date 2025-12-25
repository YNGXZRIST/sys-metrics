package methods

import (
	"errors"
	"strconv"
	"sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
)

func Update(metricType, name, value string) error {
	switch metricType {
	case "counter":
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return updateCounter(name, parsed)
	case "gauge":
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		return updateGauge(name, parsed)
	default:
		return svc.ErrUnknownMetricType
	}
}

func updateCounter(name string, value int64) error {
	storage := svc.Counters()
	counter, err := storage.Get(name)
	if errors.Is(err, memstorage.ErrNotFound) {
		counter = metrics.NewCounter(name)
	} else if err != nil {
		return err
	}
	counter.SetValue(value)
	return storage.Set(name, counter)
}

func updateGauge(name string, value float64) error {
	storage := svc.Gauges()
	gauge, err := storage.Get(name)
	if err != nil {
		gauge = metrics.NewGauge(name)
	}
	gauge.SetValue(value)
	return storage.Set(name, gauge)
}
