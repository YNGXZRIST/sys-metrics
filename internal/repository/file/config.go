package file

import (
	"fmt"
	"os"
	"path"
	"sys-metrics/internal/common"
	lgr "sys-metrics/internal/logger"
	"sys-metrics/internal/repository/metricsiface"
	"time"

	"go.uber.org/zap"
)

const DefaultFileName = "backups.metrics"

type Config struct {
	Interval       time.Duration
	StoragePath    string
	filePath       string
	Enabled        bool
	Logger         *zap.Logger
	Writer         *Writer
	Reader         *Reader
	MetricsHandler metricsiface.Handler
}

func NewConfig(mode string, storagePath string, interval time.Duration, enabled bool) (*Config, error) {
	logger, err := lgr.Initialize(mode, common.TypeBackups)
	if err != nil {
		return nil, err
	}

	var filePath string
	if mode == common.TypeModeTest {
		tmpDir, err := os.MkdirTemp("", "backups_test_dir_*")
		if err != nil {
			return nil, err
		}
		storagePath = tmpDir
		tmpFile, err := os.CreateTemp(storagePath, "backups_test_*.metrics")
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return nil, err
		}
		filePath = tmpFile.Name()
		_ = tmpFile.Close()
	} else {
		filePath = path.Join(storagePath, DefaultFileName)
	}
	err = os.MkdirAll(storagePath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	writer, err := newBackupWriter(filePath)
	if err != nil {
		return nil, err
	}

	reader, err := NewBackupReader(filePath)
	if err != nil {
		_ = writer.Close()
		return nil, err
	}

	return &Config{
		StoragePath:    storagePath,
		filePath:       filePath,
		Interval:       interval,
		Enabled:        enabled,
		Logger:         logger,
		Writer:         writer,
		Reader:         reader,
		MetricsHandler: NewFileMetricsBackupHandler(reader, writer),
	}, nil
}

func (bc *Config) NeedSync() bool {
	return bc.Interval == 0*time.Second
}

func (bc *Config) Close() error {
	var errs error
	if err := bc.Writer.Close(); err != nil {
		errs = err
	}
	if err := bc.Reader.Close(); err != nil {
		if errs != nil {
			return fmt.Errorf("multiple errors: %v; %v", errs, err)
		}
		errs = err
	}
	return errs
}
func (bc *Config) Cleanup() error {
	if bc.Writer != nil {
		_ = bc.Writer.Close()
	}
	if bc.Reader != nil {
		_ = bc.Reader.Close()
	}
	if bc.filePath != path.Join(bc.StoragePath, DefaultFileName) {
		return os.RemoveAll(bc.StoragePath)
	}
	return nil
}
func (bc *Config) getBackupFilename() string {
	return bc.filePath
}
