package metrics

import (
	"context"
	"errors"
	"fmt"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/pkg/storage"
)

type CounterBackupStorage struct {
	*storage.MemStorage[string, *metrics.Counter]
	Config         metricsiface.BackupConfig
	MetricsHandler metricsiface.Handler
}

func (s *CounterBackupStorage) Set(ctx context.Context, key string, value *metrics.Counter) error {
	if value == nil {
		return errors.New("value is nil")
	}
	if err := s.MemStorage.Set(ctx, key, value); err != nil {
		return fmt.Errorf("error setting metrics: %w", err)
	}

	if s.Config.NeedSync() {

		if err := s.MetricsHandler.Upsert(ctx, &value.Metrics); err != nil {
			return fmt.Errorf("error updating metrics: %w", err)
		}
	}

	return nil
}

func (s *CounterBackupStorage) NeedSync() bool {
	return s.Config.NeedSync()
}
