package file

import (
	"errors"
	"fmt"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/pkg/storage"
)

type gaugeBackupStorage struct {
	*storage.MemStorage[string, *metrics.Gauge]
	config         metricsiface.BackupConfig
	metricsHandler metricsiface.Handler
}

func (s *gaugeBackupStorage) Set(key string, value *metrics.Gauge) error {
	if value == nil {
		return errors.New("value is nil")
	}
	if err := s.MemStorage.Set(key, value); err != nil {
		return fmt.Errorf("error setting metrics: %w", err)
	}

	if s.config.NeedSync() {
		if err := s.metricsHandler.Upsert(&value.Metrics); err != nil {
			return fmt.Errorf("error updating metrics: %w", err)
		}
	}

	return nil
}

func (s *gaugeBackupStorage) NeedSync() bool {
	return s.config.NeedSync()
}
