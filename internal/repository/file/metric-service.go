package file

import (
	"context"
	"fmt"
	"sync"
	"sys-metrics/internal/common"
	"sys-metrics/internal/errors/labelerrors"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/internal/repository/rollback"
	"sys-metrics/internal/repository/utils"
	"time"
)

type BackupService struct {
	BackupStorage *MetricBackupStorage
	mu            sync.Mutex
}

func (s *BackupService) WriteBatchMetrics(ctx context.Context, m []models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot := s.GetAllMetricsLocked(ctx)
	byID, err := utils.ApplyBatchToStorages(ctx, m, s.Gauges(), s.Counters())
	if err != nil {
		rollback.Memory(ctx, s.Gauges(), s.Counters(), snapshot, m)
		return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to apply batch: %w", err))
	}
	updateMetrics := make([]models.Metrics, 0, len(byID))
	for _, v := range byID {
		updateMetrics = append(updateMetrics, v)
	}
	err = s.WriteBackupLocked(ctx)
	if err != nil {
		rollback.Memory(ctx, s.Gauges(), s.Counters(), snapshot, updateMetrics)
		return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to write batch: %w", err))
	}
	return nil
}

func NewBackupService(bs *MetricBackupStorage) *BackupService {
	return &BackupService{
		BackupStorage: bs,
	}
}
func (s *BackupService) GetAllMetricsLocked(ctx context.Context) []models.Metrics {
	counters := s.BackupStorage.Counters().All(ctx)
	gauges := s.BackupStorage.Gauges().All(ctx)
	var m = make([]models.Metrics, 0, len(counters)+len(gauges))
	for _, counter := range counters {
		m = append(m, counter.Metrics)
	}
	for _, gauge := range gauges {
		m = append(m, gauge.Metrics)
	}
	return m
}
func (s *BackupService) GetAllMetrics(ctx context.Context) []models.Metrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.GetAllMetricsLocked(ctx)
}
func (s *BackupService) Close(ctx context.Context) error {
	_ = ctx
	return s.BackupStorage.Close()
}
func (s *BackupService) Gauges() metricsiface.MetricStorage[*models.Gauge] {
	return s.BackupStorage.Gauges()
}

func (s *BackupService) Counters() metricsiface.MetricStorage[*models.Counter] {
	return s.BackupStorage.Counters()
}

func (s *BackupService) BackupGauges(ctx context.Context) metricsiface.BackupMetricStorage[*models.Gauge] {
	_ = ctx
	return s.BackupStorage.Gauges()
}

func (s *BackupService) BackupCounters(ctx context.Context) metricsiface.BackupMetricStorage[*models.Counter] {
	_ = ctx
	return s.BackupStorage.Counters()
}
func (s *BackupService) WriteBackupLocked(ctx context.Context) error {
	if err := s.BackupStorage.Reader.Reset(); err != nil {
		return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to reset metrics: %w", err))
	}
	metricsToWrite := s.GetAllMetricsLocked(ctx)
	if err := s.BackupStorage.Writer.file.Truncate(0); err != nil {
		return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to truncate file: %w", err))
	}
	if _, err := s.BackupStorage.Writer.file.Seek(0, 0); err != nil {
		return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to seek file: %w", err))
	}
	s.BackupStorage.Writer.writer.Reset(s.BackupStorage.Writer.file)

	return s.BackupStorage.MetricsHandler.WriteBatch(ctx, metricsToWrite)
}
func (s *BackupService) WriteBackup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.WriteBackupLocked(ctx)
}

func (s *BackupService) ReadBackup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.BackupStorage.Reader.Reset(); err != nil {
		return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to reset metrics: %w", err))
	}

	metricsData, err := s.BackupStorage.MetricsHandler.Read(ctx)
	if err != nil {
		return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to read metrics: %w", err))
	}

	for _, metric := range metricsData {
		switch metric.MType {
		case common.Counter:
			counter := &models.Counter{Metrics: metric}
			if err := s.BackupStorage.Counters().Set(ctx, metric.ID, counter); err != nil {
				return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to write metric: %w", err))
			}
		case common.Gauge:
			gauge := &models.Gauge{Metrics: metric}
			if err := s.BackupStorage.Gauges().Set(ctx, metric.ID, gauge); err != nil {
				return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to write metric: %w", err))
			}
		default:
			continue
		}
	}
	return nil
}

func (s *BackupService) InitRoutine(ctx context.Context) error {
	cfg := s.BackupStorage.Config
	if cfg.Enabled {
		if err := s.ReadBackup(ctx); err != nil {
			return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to read metrics: %w", err))
		}
	}
	if !cfg.NeedSync() {
		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := s.WriteBackup(ctx); err != nil {
					return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to write metrics: %w", err))
				}
			}
		}
	}
	return nil
}
