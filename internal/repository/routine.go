package repository

import (
	"context"
	"time"
)

func (bc *Config) InitBackupRoutine(ctx context.Context) error {
	if bc.Enabled {
		err := GetService().ReadBackup()
		if err != nil {
			return err
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
					return err
				}
			}
		}
	}

	return nil
}
