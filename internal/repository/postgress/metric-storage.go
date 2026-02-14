package postgress

import (
	"context"
	"fmt"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/model/metrics"
	m "sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/repository/metricsiface"
)

type MetricStorage struct {
	*m.BackupStorage
	Config *Config
}

func NewMetricStorage(db *db.DB) *MetricStorage {
	config := NewConfig(db)
	storage := m.NewBackupStorage(config, config.handler)
	ms := &MetricStorage{Config: config, BackupStorage: storage}
	return ms
}

func (s *MetricStorage) Close(ctx context.Context) error {
	return s.Config.conn.Close()
}

func (s *MetricStorage) GetAllMetrics(ctx context.Context) []metrics.Metrics {
	var m []metrics.Metrics
	for _, counter := range s.BackupStorage.Counters().All(ctx) {
		m = append(m, counter.Metrics)
	}
	for _, gauge := range s.BackupStorage.Gauges().All(ctx) {
		m = append(m, gauge.Metrics)
	}
	return m
}

func (s *MetricStorage) Gauges() metricsiface.MetricStorage[*metrics.Gauge] {
	return s.BackupStorage.Gauges()
}

func (s *MetricStorage) Counters() metricsiface.MetricStorage[*metrics.Counter] {
	return s.BackupStorage.Counters()
}

func (s *MetricStorage) ReadBackup(ctx context.Context) error {
	allMetrics, err := s.Config.handler.Read(ctx)
	if err != nil {
		return fmt.Errorf("read metrics from db: %w", err)
	}
	for _, v := range allMetrics {
		var err error
		switch v.MType {
		case common.Gauge:
			gauge := metrics.NewGauge(v.ID)
			gauge.SetValue(*v.Value)
			err = s.Gauges().Set(ctx, v.ID, gauge)
		case common.Counter:
			counter := metrics.NewCounter(v.ID)
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
	allMetrics := s.GetAllMetrics(ctx)
	err := s.Config.handler.WriteBatch(ctx, allMetrics)
	if err != nil {
		return fmt.Errorf("write metrics backup error: %w", err)
	}
	return nil
}

func (s *MetricStorage) InitRoutine(ctx context.Context) error {
	return nil
}
