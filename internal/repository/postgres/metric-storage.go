// Package postgres implements metricsiface.ServiceInterface on top of PostgreSQL and BackupStorage.
package postgres

import (
	"context"
	"fmt"
	"sync"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/errors/labelerrors"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/internal/repository/rollback"
	"sys-metrics/internal/repository/utils"
)

// MetricStorage combines in-memory BackupStorage with DB writes and rollback on errors.
type MetricStorage struct {
	*metrics.BackupStorage
	Config *Config
	mu     sync.Mutex
}

func (s *MetricStorage) WriteBatchMetrics(ctx context.Context, m []models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot := s.GetAllMetricsLocked(ctx)
	byID, err := utils.ApplyBatchToStorages(ctx, m, s.Gauges(), s.Counters())
	if err != nil {
		rollback.Memory(ctx, s.Gauges(), s.Counters(), snapshot, m)
		return labelerrors.NewLabelError("DB STORAGE", fmt.Errorf("failed to rollback: %w", err))
	}
	updateMetrics := make([]models.Metrics, 0, len(byID))
	for _, v := range byID {
		updateMetrics = append(updateMetrics, v)
	}
	err = s.Config.handler.WriteBatch(ctx, updateMetrics)
	if err != nil {
		rollback.Memory(ctx, s.Gauges(), s.Counters(), snapshot, updateMetrics)
		return labelerrors.NewLabelError("DB STORAGE", fmt.Errorf("failed to commit: %w", err))
	}
	return nil
}

// NewMetricStorage builds storage with a SQL handler backed by db.
func NewMetricStorage(db *db.DB) *MetricStorage {
	config := NewConfig(db)
	storage := metrics.NewBackupStorage(config, config.handler)
	ms := &MetricStorage{Config: config, BackupStorage: storage}
	return ms
}
func (s *MetricStorage) Close(ctx context.Context) error {
	_ = ctx
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
		return labelerrors.NewLabelError("DB STORAGE", fmt.Errorf("failed to read metrics: %w", err))
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
			return labelerrors.NewLabelError("DB STORAGE", fmt.Errorf("failed to set metrics: %w", err))
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
		return labelerrors.NewLabelError("DB STORAGE", fmt.Errorf("failed to batch: %w", err))
	}
	return nil
}

func (s *MetricStorage) InitRoutine(ctx context.Context) error {
	_ = ctx
	return nil
}
