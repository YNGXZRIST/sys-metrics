package memory

import (
	"context"
	"sys-metrics/internal/model/metrics"
)
import "sys-metrics/internal/repository/metricsiface"

type Service struct {
	counters metricsiface.MetricStorage[*metrics.Counter]
	gauges   metricsiface.MetricStorage[*metrics.Gauge]
}

func (s *Service) InitRoutine(ctx context.Context) error {
	return nil
}

func NewService(counters metricsiface.MetricStorage[*metrics.Counter], gauge metricsiface.MetricStorage[*metrics.Gauge]) *Service {
	return &Service{counters: counters, gauges: gauge}
}

func (s *Service) GetAllMetrics() []metrics.Metrics {
	var m []metrics.Metrics
	for _, counter := range s.counters.All() {
		m = append(m, counter.Metrics)
	}
	for _, gauge := range s.gauges.All() {
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

func (s *Service) ReadBackup() error {
	return nil
}

func (s *Service) WriteBackup() error {
	return nil
}
