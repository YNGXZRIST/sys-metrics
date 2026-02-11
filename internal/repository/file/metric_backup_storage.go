package file

import "errors"

type MetricBackupStorage struct {
	*BackupStorage
	*Config
}

func NewMetricFileBackupStorage(config *Config) (*MetricBackupStorage, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	fileBackupStorage := NewFileBackupStorage(config, config.MetricsHandler)
	metricStorage := &MetricBackupStorage{
		BackupStorage: fileBackupStorage,
		Config:        config,
	}

	return metricStorage, nil
}

func (s *MetricBackupStorage) NeedSync() bool {
	return s.Config.NeedSync()
}
