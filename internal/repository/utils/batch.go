package utils

import (
	"context"
	"fmt"
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
)

func ApplyGauge(
	ctx context.Context,
	v models.Metrics,
	gauges metricsiface.MetricStorage[*models.Gauge],
) error {
	return gauges.Set(ctx, v.ID, &models.Gauge{Metrics: v})
}

func ApplyCounter(
	ctx context.Context,
	v models.Metrics,
	counters metricsiface.MetricStorage[*models.Counter],
) (models.Metrics, error) {
	var zero models.Metrics
	existing, getErr := counters.Get(ctx, v.ID)
	if getErr != nil || existing == nil {
		c := &models.Counter{Metrics: models.Metrics{ID: v.ID, MType: common.Counter}}
		if v.Delta != nil {
			c.SetValue(*v.Delta)
		}
		if err := counters.Set(ctx, v.ID, c); err != nil {
			return zero, fmt.Errorf("write metrics %v: %w", v.MType, err)
		}
		current, _ := counters.Get(ctx, v.ID)
		if current != nil {
			return current.Metrics, nil
		}
		return zero, nil
	}
	if v.Delta != nil {
		existing.SetValue(*v.Delta)
	}
	current, _ := counters.Get(ctx, v.ID)
	if current != nil {
		return current.Metrics, nil
	}
	return zero, nil
}

func ApplyBatchToStorages(
	ctx context.Context,
	m []models.Metrics,
	gauges metricsiface.MetricStorage[*models.Gauge],
	counters metricsiface.MetricStorage[*models.Counter],
) (map[string]models.Metrics, error) {
	byID := make(map[string]models.Metrics, len(m))
	for _, v := range m {
		switch v.MType {
		case common.Gauge:
			if err := ApplyGauge(ctx, v, gauges); err != nil {
				return nil, fmt.Errorf("write metrics %v: %w", v.MType, err)
			}
			byID[v.ID] = v
		case common.Counter:
			metrics, err := ApplyCounter(ctx, v, counters)
			if err != nil {
				return nil, err
			}
			if metrics.ID != "" {
				byID[v.ID] = metrics
			}
		default:
			continue
		}
	}
	return byID, nil
}
