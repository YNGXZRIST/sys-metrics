package repository

import (
	"context"
	"fmt"
	"time"
)

func (bc *Config) InitBackupRoutine(ctx context.Context) error {
	if bc.Enabled {
		err := GetService().ReadBackup()
		if err != nil {
			return fmt.Errorf("error reading backup: %w", err)
		}
	}
	if !bc.NeedSync() {
		ticker := time.NewTicker(bc.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				err := GetService().WriteBackup()
				if err != nil {
					return fmt.Errorf("error writing backup: %w", err)
				}
			}
		}
	}

	return nil
}
