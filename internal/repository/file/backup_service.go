package file

import (
	"context"
	"fmt"
	"sync"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"time"
)

type BackupService struct {
	BackupStorage *MetricBackupStorage
	mu            sync.Mutex
}

func NewBackupService(bs *MetricBackupStorage) *BackupService {
	return &BackupService{
		BackupStorage: bs,
	}
}

func (s *BackupService) GetAllMetrics() []metrics.Metrics {
	var m []metrics.Metrics
	for _, counter := range s.BackupStorage.Counters().All() {
		m = append(m, counter.Metrics)
	}
	for _, gauge := range s.BackupStorage.Gauges().All() {
		m = append(m, gauge.Metrics)
	}
	return m
}

func (s *BackupService) Gauges() metricsiface.MetricStorage[*metrics.Gauge] {
	return s.BackupStorage.Gauges()
}

func (s *BackupService) Counters() metricsiface.MetricStorage[*metrics.Counter] {
	return s.BackupStorage.Counters()
}

func (s *BackupService) BackupGauges() metricsiface.BackupMetricStorage[*metrics.Gauge] {
	return s.BackupStorage.Gauges()
}

func (s *BackupService) BackupCounters() metricsiface.BackupMetricStorage[*metrics.Counter] {
	return s.BackupStorage.Counters()
}

func (s *BackupService) WriteBackup() error {
	if err := s.BackupStorage.Reader.Reset(); err != nil {
		return fmt.Errorf("error writing reset backup: %w", err)
	}
	metricsToWrite := s.GetAllMetrics()
	if err := s.BackupStorage.Writer.file.Truncate(0); err != nil {
		return fmt.Errorf("error writing truncate backup: %w", err)
	}
	if _, err := s.BackupStorage.Writer.file.Seek(0, 0); err != nil {
		return fmt.Errorf("error writing seek backup: %w", err)
	}
	s.BackupStorage.Writer.writer.Reset(s.BackupStorage.Writer.file)

	return s.BackupStorage.MetricsHandler.WriteBatch(metricsToWrite)
}

func (s *BackupService) ReadBackup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.BackupStorage.Reader.Reset(); err != nil {
		return fmt.Errorf("error writing reset backup: %w", err)
	}

	metricsData, err := s.BackupStorage.MetricsHandler.Read()
	if err != nil {
		return fmt.Errorf("error writing read backup: %w", err)
	}

	for _, metric := range metricsData {
		switch metric.MType {
		case common.Counter:
			counter := &metrics.Counter{Metrics: metric}
			if err := s.BackupStorage.Counters().Set(metric.ID, counter); err != nil {
				return fmt.Errorf("error writing counter: %w", err)
			}
		case common.Gauge:
			gauge := &metrics.Gauge{Metrics: metric}
			if err := s.BackupStorage.Gauges().Set(metric.ID, gauge); err != nil {
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
		if err := s.ReadBackup(); err != nil {
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
				if err := s.WriteBackup(); err != nil {
					return fmt.Errorf("error writing backup: %w", err)
				}
			}
		}
	}
	return nil
}
