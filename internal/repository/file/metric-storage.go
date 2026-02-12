package file

import (
	"errors"
	"fmt"
	m "sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/repository/metricsiface"
)

type MetricBackupStorage struct {
	*m.BackupStorage
	*Config
	Writer         *Writer
	Reader         *Reader
	MetricsHandler metricsiface.Handler
}

func NewMetricFileBackupStorage(config *Config) (*MetricBackupStorage, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	filePath := config.getBackupFilename()
	writer, err := newBackupWriter(filePath)
	if err != nil {
		return nil, fmt.Errorf("new backup writer: %w", err)
	}
	reader, err := NewBackupReader(filePath)
	if err != nil {
		_ = writer.Close()
		return nil, fmt.Errorf("new backup reader: %w", err)
	}
	handler := NewFileMetricsBackupHandler(reader, writer)
	backupStorage := m.NewBackupStorage(config, handler)
	return &MetricBackupStorage{
		BackupStorage:  backupStorage,
		Config:         config,
		Writer:         writer,
		Reader:         reader,
		MetricsHandler: handler,
	}, nil
}

func (s *MetricBackupStorage) NeedSync() bool {
	return s.Config.NeedSync()
}

func (s *MetricBackupStorage) Close() error {
	var errs error
	if err := s.Writer.Close(); err != nil {
		errs = err
	}
	if err := s.Reader.Close(); err != nil {
		if errs != nil {
			return fmt.Errorf("multiple errors: %v; %v", errs, err)
		}
		errs = err
	}
	return errs
}
