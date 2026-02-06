package repository

import (
	"fmt"
	"os"
	"path"
	"sys-metrics/internal/common"
	lgr "sys-metrics/internal/logger"
	"sys-metrics/pkg/filesystem"
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
	MetricsHandler MetricsBackupHandler
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

	err = filesystem.CreateDirIfNotExists(storagePath)
	if err != nil {
		return nil, err
	}
	writer, err := NewBackupWriter(filePath)
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
