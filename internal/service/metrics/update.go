package metrics

import (
	"errors"
	"fmt"
	"strconv"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository"
	"sys-metrics/pkg/storage"
)

func Update(metricType, name, value string) error {
	switch metricType {
	case common.Counter:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid counter metric value: %w", err)
		}
		return updateCounter(name, parsed)
	case common.Gauge:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid gauge metric value: %w", err)
		}
		return updateGauge(name, parsed)
	default:
		return ErrUnknownMetricType
	}
}

func updateCounter(name string, value int64) error {
	repo := repository.Counters()
	counter, err := repo.Get(name)
	if errors.Is(err, storage.ErrNotFound) {
		counter = metrics.NewCounter(name)
	} else if err != nil {
		return fmt.Errorf("repository error: %w", err)
	}
	counter.SetValue(value)
	err = repo.Set(name, counter)
	if err != nil {
		return fmt.Errorf("repository error set counter: %w", err)
	}
	return nil
}

func updateGauge(name string, value float64) error {
	repo := repository.Gauges()
	gauge, err := repo.Get(name)
	if err != nil {
		gauge = metrics.NewGauge(name)
	}
	gauge.SetValue(value)
	err = repo.Set(name, gauge)
	if err != nil {
		return fmt.Errorf("repository error set gauge: %w", err)
	}
	return nil
}
