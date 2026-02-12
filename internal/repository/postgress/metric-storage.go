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
	rows, err := s.Config.conn.QueryContext(ctx, "SELECT * from metrics")
	if err != nil {
		return err
	}
	for rows.Next() {
		var metric metrics.Metrics
		err = rows.Scan(metric)
		if err != nil {
			return fmt.Errorf("failed to scan metrics: %v", err)
		}
		switch metric.MType {
		case common.Gauge:
			gauge := metrics.NewGauge(metric.ID)
			gauge.SetValue(*metric.Value)
		case common.Counter:
			counter := metrics.NewCounter(metric.ID)
			counter.SetValue(*metric.Delta)
			err := s.Counters().Set(ctx, metric.ID, counter)
			if err != nil {
				return fmt.Errorf("metric error '%s': %w", metric.ID, err)
			}
		}
	}
	//TODO implement me
	return nil
}

func (s *MetricStorage) WriteBackup(ctx context.Context) error {
	//TODO implement me
	return nil
}

func (s *MetricStorage) InitRoutine(ctx context.Context) error {
	return nil
}
