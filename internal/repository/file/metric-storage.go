package file

import (
	"fmt"
	"sys-metrics/internal/errors/labelerrors"
	m "sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/repository/metricsiface"
)

// generate:reset

type MetricBackupStorage struct {
	*m.BackupStorage
	*Config
	Writer         *Writer
	Reader         *Reader
	MetricsHandler metricsiface.Handler
}

func NewMetricFileBackupStorage(config *Config) (*MetricBackupStorage, error) {
	if config == nil {
		return nil, labelerrors.NewLabelError("BACKUP", fmt.Errorf("config is nil"))
	}
	filePath := config.getBackupFilename()
	writer, err := newBackupWriter(filePath)
	if err != nil {
		return nil, labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to create backup writer: %w", err))
	}
	reader, err := NewBackupReader(filePath)
	if err != nil {
		_ = writer.Close()
		return nil, labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to create backup reader: %w", err))
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
			return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to close reader: %w", err))
		}
		errs = err
	}
	return errs
}
