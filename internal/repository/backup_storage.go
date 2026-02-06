package repository

import (
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
)

type BackupConfig interface {
	NeedSync() bool
}

func NewFileBackupStorage(config BackupConfig, metricsHandler MetricsBackupHandler) *FileBackupStorage {
	return &FileBackupStorage{
		counters:       storage.NewMemStorage[string, *metrics.Counter](),
		gauges:         storage.NewMemStorage[string, *metrics.Gauge](),
		config:         config,
		metricsHandler: metricsHandler,
	}
}

func (s *FileBackupStorage) Counters() BackupMetricStorage[*metrics.Counter] {
	return &counterBackupStorage{
		MemStorage:     s.counters,
		config:         s.config,
		metricsHandler: s.metricsHandler,
	}
}

func (s *FileBackupStorage) Gauges() BackupMetricStorage[*metrics.Gauge] {
	return &gaugeBackupStorage{
		MemStorage:     s.gauges,
		config:         s.config,
		metricsHandler: s.metricsHandler,
	}
}

type counterBackupStorage struct {
	*storage.MemStorage[string, *metrics.Counter]
	config         BackupConfig
	metricsHandler MetricsBackupHandler
}

func (s *counterBackupStorage) Set(key string, value *metrics.Counter) error {
	if err := s.MemStorage.Set(key, value); err != nil {
		return err
	}

	if s.config.NeedSync() {
		if err := s.metricsHandler.Upsert(&value.Metrics); err != nil {
			return err
		}
	}

	return nil
}

func (s *counterBackupStorage) NeedSync() bool {
	return s.config.NeedSync()
}

type gaugeBackupStorage struct {
	*storage.MemStorage[string, *metrics.Gauge]
	config         BackupConfig
	metricsHandler MetricsBackupHandler
}

func (s *gaugeBackupStorage) Set(key string, value *metrics.Gauge) error {
	if err := s.MemStorage.Set(key, value); err != nil {
		return err
	}

	if s.config.NeedSync() {
		if err := s.metricsHandler.Upsert(&value.Metrics); err != nil {
			return err
		}
	}

	return nil
}

func (s *gaugeBackupStorage) NeedSync() bool {
	return s.config.NeedSync()
}
