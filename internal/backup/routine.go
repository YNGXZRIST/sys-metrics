package backup

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	svc "sys-metrics/internal/service/metrics"
	"time"
)

func (bc *BackupConfig) InitMetricsFromBackup() error {
	m, err := bc.Reader.ReadFromBackup()
	if err != nil {
		return err
	}
	for _, v := range m {
		switch v.MType {
		case common.Counter:
			counterMetric := metrics.NewCounter(v.ID)
			counterMetric.SetValue(*v.Delta)
			err := svc.Counters().Set(v.ID, counterMetric)
			if err != nil {
				return err
			}
		case common.Gauge:
			gaugeMetric := metrics.NewGauge(v.ID)
			gaugeMetric.SetValue(*v.Value)
			err := svc.Gauges().Set(v.ID, gaugeMetric)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (bc *BackupConfig) InitBackupRoutine(ctx context.Context) error {
	if bc.Enabled {
		err := bc.InitMetricsFromBackup()
		if err != nil {
			return err
		}
	}
	if !bc.IsSyncBackup() {
		ticker := time.NewTicker(bc.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				err := bc.UpsertBatchMetricsToBackup(svc.GetAllMetrics())
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
