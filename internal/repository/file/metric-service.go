package file

import (
	"context"
	"fmt"
	"sync"
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/internal/repository/rollback"
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
	for _, v := range m {
		var err error
		switch v.MType {
		case common.Gauge:
			err = s.Gauges().Set(ctx, v.ID, &models.Gauge{Metrics: v})
		case common.Counter:
			err = s.Counters().Set(ctx, v.ID, &models.Counter{Metrics: v})
		}
		if err != nil {
			rollback.Memory(ctx, s.Gauges(), s.Counters(), snapshot, m)
			return fmt.Errorf("write memory metrics  %v error: %w", v.MType, err)
		}
	}
	err := s.WriteBackupLocked(ctx)
	if err != nil {
		rollback.Memory(ctx, s.Gauges(), s.Counters(), snapshot, m)
		return fmt.Errorf("write backup metrics %v error: %w", m, err)
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
	return s.BackupStorage.Close()
}
func (s *BackupService) Gauges() metricsiface.MetricStorage[*models.Gauge] {
	return s.BackupStorage.Gauges()
}

func (s *BackupService) Counters() metricsiface.MetricStorage[*models.Counter] {
	return s.BackupStorage.Counters()
}

func (s *BackupService) BackupGauges(ctx context.Context) metricsiface.BackupMetricStorage[*models.Gauge] {
	return s.BackupStorage.Gauges()
}

func (s *BackupService) BackupCounters(ctx context.Context) metricsiface.BackupMetricStorage[*models.Counter] {
	return s.BackupStorage.Counters()
}
func (s *BackupService) WriteBackupLocked(ctx context.Context) error {
	if err := s.BackupStorage.Reader.Reset(); err != nil {
		return fmt.Errorf("error writing reset backup: %w", err)
	}
	metricsToWrite := s.GetAllMetricsLocked(ctx)
	if err := s.BackupStorage.Writer.file.Truncate(0); err != nil {
		return fmt.Errorf("error writing truncate backup: %w", err)
	}
	if _, err := s.BackupStorage.Writer.file.Seek(0, 0); err != nil {
		return fmt.Errorf("error writing seek backup: %w", err)
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
		return fmt.Errorf("error writing reset backup: %w", err)
	}

	metricsData, err := s.BackupStorage.MetricsHandler.Read(ctx)
	if err != nil {
		return fmt.Errorf("error writing read backup: %w", err)
	}

	for _, metric := range metricsData {
		switch metric.MType {
		case common.Counter:
			counter := &models.Counter{Metrics: metric}
			if err := s.BackupStorage.Counters().Set(ctx, metric.ID, counter); err != nil {
				return fmt.Errorf("error writing counter: %w", err)
			}
		case common.Gauge:
			gauge := &models.Gauge{Metrics: metric}
			if err := s.BackupStorage.Gauges().Set(ctx, metric.ID, gauge); err != nil {
				return fmt.Errorf("error writing gauge: %w", err)
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
			return fmt.Errorf("error reading backup: %w", err)
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
					return fmt.Errorf("error writing backup: %w", err)
				}
			}
		}
	}
	return nil
}
