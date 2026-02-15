package postgress

import (
	"context"
	"fmt"
	"sync"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/internal/repository/rollback"
)

type MetricStorage struct {
	*metrics.BackupStorage
	Config *Config
	mu     sync.Mutex
}

func (s *MetricStorage) WriteBatchMetrics(ctx context.Context, m []models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot := s.GetAllMetricsLocked(ctx)
	updateMetrics := make([]models.Metrics, 0, len(m))
	for _, v := range m {
		var err error
		switch v.MType {
		case common.Gauge:
			err = s.Gauges().Set(ctx, v.ID, &models.Gauge{Metrics: v})
		case common.Counter:
			err = s.Counters().Set(ctx, v.ID, &models.Counter{Metrics: v})
		default:
			continue
		}
		if err != nil {
			rollback.Memory(ctx, s.Gauges(), s.Counters(), snapshot, m)
			return fmt.Errorf("write metrics %v error: %w", v.MType, err)
		}
		updateMetrics = append(updateMetrics, v)
	}
	err := s.Config.handler.WriteBatch(ctx, updateMetrics)
	if err != nil {
		rollback.Memory(ctx, s.Gauges(), s.Counters(), snapshot, updateMetrics)
		return fmt.Errorf("write metrics backup error: %w", err)
	}
	return nil
}
func NewMetricStorage(db *db.DB) *MetricStorage {
	config := NewConfig(db)
	storage := metrics.NewBackupStorage(config, config.handler)
	ms := &MetricStorage{Config: config, BackupStorage: storage}
	return ms
}
func (s *MetricStorage) Close(ctx context.Context) error {
	return s.Config.conn.Close()
}
func (s *MetricStorage) GetAllMetricsLocked(ctx context.Context) []models.Metrics {
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
func (s *MetricStorage) GetAllMetrics(ctx context.Context) []models.Metrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.GetAllMetricsLocked(ctx)
}

func (s *MetricStorage) Gauges() metricsiface.MetricStorage[*models.Gauge] {
	return s.BackupStorage.Gauges()
}

func (s *MetricStorage) Counters() metricsiface.MetricStorage[*models.Counter] {
	return s.BackupStorage.Counters()
}

func (s *MetricStorage) ReadBackup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	allMetrics, err := s.Config.handler.Read(ctx)
	if err != nil {
		return fmt.Errorf("read metrics from db: %w", err)
	}
	for _, v := range allMetrics {
		var err error
		switch v.MType {
		case common.Gauge:
			gauge := models.NewGauge(v.ID)
			gauge.SetValue(*v.Value)
			err = s.Gauges().Set(ctx, v.ID, gauge)
		case common.Counter:
			counter := models.NewCounter(v.ID)
			counter.SetValue(*v.Delta)
			err = s.Counters().Set(ctx, v.ID, counter)
		default:
			continue
		}
		if err != nil {
			return fmt.Errorf("set metrics error: %w", err)
		}
	}
	s.Config.initialized = true
	return nil
}
func (s *MetricStorage) WriteBackup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	allMetrics := s.GetAllMetricsLocked(ctx)
	err := s.Config.handler.WriteBatch(ctx, allMetrics)
	if err != nil {
		return fmt.Errorf("write metrics backup error: %w", err)
	}
	return nil
}

func (s *MetricStorage) InitRoutine(ctx context.Context) error {
	return nil
}
