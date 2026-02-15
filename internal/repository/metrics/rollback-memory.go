package metrics

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
)

func RollbackMemory(ctx context.Context, gauges metricsiface.MetricStorage[*metrics.Gauge], counters metricsiface.MetricStorage[*metrics.Counter], snapshot, newMetrics []metrics.Metrics) {
	gaugeOld := make(map[string]metrics.Metrics, len(snapshot))
	counterOld := make(map[string]metrics.Metrics, len(snapshot))
	for _, v := range snapshot {
		switch v.MType {
		case common.Gauge:
			gaugeOld[v.ID] = v
		case common.Counter:
			counterOld[v.ID] = v
		}
	}
	for _, v := range newMetrics {
		switch v.MType {
		case common.Gauge:
			if old, ok := gaugeOld[v.ID]; ok {
				_ = gauges.Set(ctx, v.ID, &metrics.Gauge{Metrics: old})
			} else {
				_ = gauges.Delete(ctx, v.ID)
			}
		case common.Counter:
			if old, ok := counterOld[v.ID]; ok {
				_ = counters.Set(ctx, v.ID, &metrics.Counter{Metrics: old})
			} else {
				_ = counters.Delete(ctx, v.ID)
			}
		}
	}

}
