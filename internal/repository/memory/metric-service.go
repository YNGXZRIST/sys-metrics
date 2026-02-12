package memory

import (
	"context"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
)
import "sys-metrics/internal/repository/metricsiface"

type Service struct {
	counters metricsiface.MetricStorage[*metrics.Counter]
	gauges   metricsiface.MetricStorage[*metrics.Gauge]
}

func (s *Service) InitRoutine(ctx context.Context) error {
	return nil
}

func NewService() *Service {
	counters := storage.NewMemStorage[string, *metrics.Counter]()
	gauges := storage.NewMemStorage[string, *metrics.Gauge]()
	return &Service{counters: counters, gauges: gauges}
}
func (s *Service) Close(ctx context.Context) error {
	return nil
}
func (s *Service) GetAllMetrics(ctx context.Context) []metrics.Metrics {
	var m []metrics.Metrics
	for _, counter := range s.counters.All(ctx) {
		m = append(m, counter.Metrics)
	}
	for _, gauge := range s.gauges.All(ctx) {
		m = append(m, gauge.Metrics)
	}
	return m
}
func (s *Service) Gauges() metricsiface.MetricStorage[*metrics.Gauge] {
	return s.gauges
}

func (s *Service) Counters() metricsiface.MetricStorage[*metrics.Counter] {
	return s.counters
}

func (s *Service) ReadBackup(ctx context.Context) error {
	return nil
}

func (s *Service) WriteBackup(ctx context.Context) error {
	return nil
}
