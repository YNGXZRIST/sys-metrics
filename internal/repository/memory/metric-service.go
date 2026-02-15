package memory

import (
	"context"
	"fmt"
	"sync"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"sys-metrics/internal/repository/rollback"
	"sys-metrics/internal/repository/utils"
	"sys-metrics/pkg/storage"
)

type Service struct {
	counters metricsiface.MetricStorage[*models.Counter]
	gauges   metricsiface.MetricStorage[*models.Gauge]
	mu       sync.Mutex
}

func (s *Service) WriteBatchMetrics(ctx context.Context, m []models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot := s.GetAllMetricsLocked(ctx)
	_, err := utils.ApplyBatchToStorages(ctx, m, s.Gauges(), s.Counters())
	if err != nil {
		rollback.Memory(ctx, s.Gauges(), s.Counters(), snapshot, m)
		return fmt.Errorf("write metrics: %w", err)
	}
	return nil
}

func (s *Service) InitRoutine(ctx context.Context) error {
	_ = ctx
	return nil
}

func NewService() *Service {
	counters := storage.NewMemStorage[string, *models.Counter]()
	gauges := storage.NewMemStorage[string, *models.Gauge]()
	return &Service{counters: counters, gauges: gauges}
}
func (s *Service) Close(ctx context.Context) error {
	_ = ctx
	return nil
}
func (s *Service) GetAllMetricsLocked(ctx context.Context) []models.Metrics {
	counters := s.counters.All(ctx)
	gauges := s.gauges.All(ctx)
	var m = make([]models.Metrics, 0, len(counters)+len(gauges))
	for _, counter := range counters {
		m = append(m, counter.Metrics)
	}
	for _, gauge := range gauges {
		m = append(m, gauge.Metrics)
	}
	return m
}
func (s *Service) GetAllMetrics(ctx context.Context) []models.Metrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.GetAllMetricsLocked(ctx)
	return m
}
func (s *Service) Gauges() metricsiface.MetricStorage[*models.Gauge] {
	return s.gauges
}

func (s *Service) Counters() metricsiface.MetricStorage[*models.Counter] {
	return s.counters
}

func (s *Service) ReadBackup(ctx context.Context) error {
	_ = ctx
	return nil
}

func (s *Service) WriteBackup(ctx context.Context) error {
	_ = ctx
	return nil
}
