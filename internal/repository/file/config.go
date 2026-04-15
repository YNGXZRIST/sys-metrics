// Package file implements file-backed metric persistence and background sync configuration.
package file

import (
	"fmt"
	"os"
	"path"
	"sys-metrics/internal/common"
	"sys-metrics/internal/errors/labelerrors"
	lgr "sys-metrics/internal/logger"
	"time"

	"go.uber.org/zap"
)

// DefaultFileName is the default backup file name under StoragePath.
const DefaultFileName = "backups.metrics"

// Config describes flush interval, path, and whether file backup is enabled.
type Config struct {
	Interval    time.Duration
	StoragePath string
	filePath    string
	Enabled     bool
	Logger      *zap.Logger
}

// NewConfig prepares backup config and logger; in test mode it uses temporary files.
func NewConfig(mode string, storagePath string, interval time.Duration, enabled bool) (*Config, error) {
	logger, err := lgr.Initialize(mode, common.TypeBackups)
	if err != nil {
		return nil, labelerrors.NewLabelError("FILE", fmt.Errorf("failed to initialize logger: %w", err))
	}

	var filePath string
	if mode == common.TypeModeTest {
		tmpDir, err := os.MkdirTemp("", "backups_test_dir_*")
		if err != nil {
			return nil, labelerrors.NewLabelError("FILE", fmt.Errorf("failed to create temp dir: %w", err))
		}
		storagePath = tmpDir
		tmpFile, err := os.CreateTemp(storagePath, "backups_test_*.metrics")
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return nil, labelerrors.NewLabelError("FILE", fmt.Errorf("cannot create temp backups.metrics file: %w", err))
		}
		filePath = tmpFile.Name()
		_ = tmpFile.Close()
	} else {
		filePath = path.Join(storagePath, DefaultFileName)
	}
	if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
		return nil, labelerrors.NewLabelError("FILE", fmt.Errorf("failed to create storage path: %w", err))
	}

	return &Config{
		StoragePath: storagePath,
		filePath:    filePath,
		Interval:    interval,
		Enabled:     enabled,
		Logger:      logger,
	}, nil
}

func (bc *Config) NeedSync() bool {
	return bc.Interval == 0*time.Second
}

func (bc *Config) Close() error {
	return nil
}

func (bc *Config) Cleanup() error {
	if bc.filePath != path.Join(bc.StoragePath, DefaultFileName) {
		return os.RemoveAll(bc.StoragePath)
	}
	return nil
}

func (bc *Config) getBackupFilename() string {
	return bc.filePath
}
