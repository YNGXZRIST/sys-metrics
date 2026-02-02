package metrics

import (
	"errors"
	"strconv"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/memstorage"
)

func Update(metricType, name, value string, isNeedBackup bool) error {
	switch metricType {
	case common.Counter:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return updateCounter(name, parsed, isNeedBackup)
	case common.Gauge:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		return updateGauge(name, parsed, isNeedBackup)
	default:
		return ErrUnknownMetricType
	}
}

func updateCounter(name string, value int64, isNeedBackup bool) error {
	storage := Counters()
	counter, err := storage.Get(name)
	if errors.Is(err, memstorage.ErrNotFound) {
		counter = metrics.NewCounter(name)
	} else if err != nil {
		return err
	}
	counter.SetValue(value)
	if isNeedBackup {

	}
	return storage.Set(name, counter)
}

func updateGauge(name string, value float64, isNeedBackup bool) error {
	storage := Gauges()
	gauge, err := storage.Get(name)
	if err != nil {
		gauge = metrics.NewGauge(name)
	}
	gauge.SetValue(value)
	return storage.Set(name, gauge)
}
