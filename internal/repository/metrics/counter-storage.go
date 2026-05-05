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

type CounterBackupStorage struct {
	*storage.MemStorage[string, *metrics.Counter]
	Config         metricsiface.BackupConfig
	MetricsHandler metricsiface.Handler
}

func (s *CounterBackupStorage) Set(ctx context.Context, key string, value *metrics.Counter) error {
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

func (s *CounterBackupStorage) NeedSync() bool {
	return s.Config.NeedSync()
}
