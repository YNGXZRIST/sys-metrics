package metrics

import (
	"context"
	"errors"
	"fmt"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/pkg/storage"
)

type GaugeBackupStorage struct {
	*storage.MemStorage[string, *metrics.Gauge]
	Config         metricsiface.BackupConfig
	MetricsHandler metricsiface.Handler
}

func (s *GaugeBackupStorage) Set(ctx context.Context, key string, value *metrics.Gauge) error {
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

func (s *GaugeBackupStorage) NeedSync() bool {
	return s.Config.NeedSync()
}
