package repository

type MetricBackupStorage struct {
	*FileBackupStorage
	*Config
}

func NewMetricBackupStorage(config *Config) (*MetricBackupStorage, error) {
	fileBackupStorage := NewFileBackupStorage(config, config.MetricsHandler)
	metricStorage := &MetricBackupStorage{
		FileBackupStorage: fileBackupStorage,
		Config:            config,
	}

	return metricStorage, nil
}

func (s *MetricBackupStorage) NeedSync() bool {
	return s.Config.NeedSync()
}
