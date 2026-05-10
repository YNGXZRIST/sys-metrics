package metrics

import (
	"context"
	"fmt"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/pkg/storage"
)

// generate:reset

type GaugeBackupStorage struct {
	*storage.MemStorage[string, *metrics.Gauge]
	Config         metricsiface.BackupConfig
	MetricsHandler metricsiface.Handler
}

func (s *GaugeBackupStorage) Set(ctx context.Context, key string, value *metrics.Gauge) error {
	if value == nil {
		return labelerrors.NewLabelError("METRICS", fmt.Errorf("value is nil"))
	}
	if err := s.MemStorage.Set(ctx, key, value); err != nil {
		return labelerrors.NewLabelError("METRICS", fmt.Errorf("failed to set metrics: %w", err))
	}

	if s.Config.NeedSync() {
		if err := s.MetricsHandler.Upsert(ctx, &value.Metrics); err != nil {
			return labelerrors.NewLabelError("METRICS", fmt.Errorf("failed to sync metrics: %w", err))
		}
	}

	return nil
}

func (s *GaugeBackupStorage) NeedSync() bool {
	return s.Config.NeedSync()
}
