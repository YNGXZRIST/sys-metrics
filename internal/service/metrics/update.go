package metrics

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	storage "sys-metrics/internal/repository/metrics"
	mem "sys-metrics/pkg/storage"
)

func Update(ctx context.Context, metricType, name, value string) error {
	switch metricType {
	case common.Counter:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid counter metric value: %w", err)
		}
		return updateCounter(ctx, name, parsed)
	case common.Gauge:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid gauge metric value: %w", err)
		}
		return updateGauge(ctx, name, parsed)
	default:
		return ErrUnknownMetricType
	}
}

func updateCounter(ctx context.Context, name string, value int64) error {
	repo := storage.Counters()
	counter, err := repo.Get(ctx, name)
	if errors.Is(err, mem.ErrNotFound) {
		counter = models.NewCounter(name)
	} else if err != nil {
		return fmt.Errorf("repository error: %w", err)
	}
	counter.SetValue(value)
	err = repo.Set(ctx, name, counter)
	if err != nil {
		return fmt.Errorf("repository error set counter: %w", err)
	}
	return nil
}

func updateGauge(ctx context.Context, name string, value float64) error {
	repo := storage.Gauges()
	gauge, err := repo.Get(ctx, name)
	if err != nil {
		gauge = models.NewGauge(name)
	}
	gauge.SetValue(value)
	err = repo.Set(ctx, name, gauge)
	if err != nil {
		return fmt.Errorf("repository error set gauge: %w", err)
	}
	return nil
}
func BatchUpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	err := storage.WriteBatchMetrics(ctx, metrics)
	if err != nil {
		return fmt.Errorf("write batch metrics: %w", err)
	}
	return nil
}
