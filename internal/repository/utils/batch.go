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
	existing, err := gauges.Get(ctx, v.ID)
	if err == nil && existing != nil {
		// Update existing gauge in-place to avoid allocating a new object.
		if v.Value != nil {
			existing.SetValue(*v.Value)
		} else {
			existing.SetValue(0)
		}
		return gauges.Set(ctx, v.ID, existing)
	}

	g := models.NewGauge(v.ID)
	if v.Value != nil {
		g.SetValue(*v.Value)
	} else {
		g.SetValue(0)
	}
	return gauges.Set(ctx, v.ID, g)
}

func ApplyCounter(
	ctx context.Context,
	v models.Metrics,
	counters metricsiface.MetricStorage[*models.Counter],
) (models.Metrics, error) {
	var zero models.Metrics
	existing, getErr := counters.Get(ctx, v.ID)
	if getErr != nil || existing == nil {
		c := models.NewCounter(v.ID)
		if v.Delta != nil {
			c.SetValue(*v.Delta)
		}
		if err := counters.Set(ctx, v.ID, c); err != nil {
			return zero, fmt.Errorf("failed to batch counter %s: %w", v.ID, err)
		}
		return c.Metrics, nil
	}
	if v.Delta != nil {
		existing.SetValue(*v.Delta)
	}
	if err := counters.Set(ctx, v.ID, existing); err != nil {
		return zero, fmt.Errorf("failed to batch counter %s: %w", v.ID, err)
	}
	return existing.Metrics, nil
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
				return nil, fmt.Errorf("failed to batch gauge %s: %w", v.ID, err)
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
