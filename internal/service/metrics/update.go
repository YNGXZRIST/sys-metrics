// Package metrics implements single-metric and batch updates on top of the global storage service.
package metrics

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sys-metrics/internal/common"
	"sys-metrics/internal/errors/labelerrors"
	models "sys-metrics/internal/model/metrics"
	storage "sys-metrics/internal/repository/metrics"
	mem "sys-metrics/pkg/storage"
)

// ErrUnknownMetricType is returned when Update receives an unknown metricType.
var (
	ErrUnknownMetricType = errors.New("unknown metric type")
	ErrMetricRead        = errors.New("metric read failed")
)

// Update parses value and updates a counter or gauge in storage by name.
func (ms *MetricService) Update(ctx context.Context, metricType, name, value string) error {
	switch metricType {
	case common.Counter:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return labelerrors.NewLabelError("UPDATE", fmt.Errorf("invalid counter metric value: %w", err))
		}
		if err := updateCounter(ctx, name, parsed); err != nil {
			return err
		}
		ms.notifyUpdated(ctx, []string{name})
		return nil
	case common.Gauge:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return labelerrors.NewLabelError("UPDATE", fmt.Errorf("invalid gauge metric value: %w", err))
		}
		if err := updateGauge(ctx, name, parsed); err != nil {
			return err
		}
		ms.notifyUpdated(ctx, []string{name})
		return nil
	default:
		return ErrUnknownMetricType
	}
}

func (ms *MetricService) UpdateMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	value, err := metricValue(metric)
	if err != nil {
		return models.Metrics{}, err
	}
	if err := ms.Update(ctx, metric.MType, metric.ID, value); err != nil {
		return models.Metrics{}, err
	}
	updated, err := ms.GetMetric(ctx, metric.MType, metric.ID)
	if err != nil {
		return models.Metrics{}, fmt.Errorf("%w: %v", ErrMetricRead, err)
	}
	return updated, nil
}

func (ms *MetricService) GetMetric(ctx context.Context, metricType, name string) (models.Metrics, error) {
	switch metricType {
	case common.Counter:
		v, err := storage.Counters().Get(ctx, name)
		if err != nil {
			return models.Metrics{}, labelerrors.NewLabelError("COUNTER", mem.ErrNotFound)
		}
		return v.Metrics, nil
	case common.Gauge:
		v, err := storage.Gauges().Get(ctx, name)
		if err != nil {
			return models.Metrics{}, labelerrors.NewLabelError("GAUGE", mem.ErrNotFound)
		}
		return v.Metrics, nil
	default:
		return models.Metrics{}, ErrUnknownMetricType
	}
}

func (ms *MetricService) MetricValue(metric models.Metrics) (string, error) {
	switch metric.MType {
	case common.Counter:
		return strconv.FormatInt(*metric.Delta, 10), nil
	case common.Gauge:
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64), nil
	default:
		return "", ErrUnknownMetricType
	}
}

func metricValue(metric models.Metrics) (string, error) {
	switch metric.MType {
	case common.Gauge:
		if metric.Value == nil {
			return "0", nil
		}
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64), nil
	case common.Counter:
		if metric.Delta == nil {
			return "0", nil
		}
		return strconv.FormatInt(*metric.Delta, 10), nil
	default:
		return "", ErrUnknownMetricType
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
		return labelerrors.NewLabelError("UPDATE", fmt.Errorf("repository error set counter: %w", err))
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
		return labelerrors.NewLabelError("UPDATE", fmt.Errorf("repository error set gauge: %w", err))
	}
	return nil
}

// BatchUpdateMetrics delegates batch persistence to the active metrics service.
func (ms *MetricService) BatchUpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	err := storage.WriteBatchMetrics(ctx, metrics)
	if err != nil {
		return labelerrors.NewLabelError("UPDATE", fmt.Errorf("write batch metrics: %w", err))
	}
	mNames := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		mNames = append(mNames, metric.ID)
	}
	ms.notifyUpdated(ctx, mNames)
	return nil
}
