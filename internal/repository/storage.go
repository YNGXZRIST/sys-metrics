package repository

import (
	"sync"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/storage"
)

var defaultService ServiceInterface

type ServiceInterface interface {
	GetAllMetrics() []metrics.Metrics
	Gauges() MetricStorage[*metrics.Gauge]
	Counters() MetricStorage[*metrics.Counter]
	ReadBackup() error
	WriteBackup() error
}

type MetricStorage[V any] interface {
	storage.Storage[string, V]
}

type BackupMetricStorage[V any] interface {
	MetricStorage[V]
	NeedSync() bool
}

type MemoryService struct {
	counters MetricStorage[*metrics.Counter]
	gauges   MetricStorage[*metrics.Gauge]
}

type BackupMemoryService struct {
	backupStorage *MetricBackupStorage
	mu            sync.Mutex
}

func Init(counters MetricStorage[*metrics.Counter], gauges MetricStorage[*metrics.Gauge]) *MemoryService {
	service := &MemoryService{
		counters: counters,
		gauges:   gauges,
	}
	defaultService = service
	return service
}

func InitBackup(storage *MetricBackupStorage) *BackupMemoryService {
	service := &BackupMemoryService{
		backupStorage: storage,
	}
	defaultService = service
	return service
}

func Counters() MetricStorage[*metrics.Counter] {
	return defaultService.Counters()
}

func Gauges() MetricStorage[*metrics.Gauge] {
	return defaultService.Gauges()
}

func GetService() ServiceInterface {
	return defaultService
}

func GetAllMetrics() []metrics.Metrics {
	return defaultService.GetAllMetrics()
}

func (s *MemoryService) GetAllMetrics() []metrics.Metrics {
	var m []metrics.Metrics
	for _, counter := range s.counters.All() {
		m = append(m, counter.Metrics)
	}
	for _, gauge := range s.gauges.All() {
		m = append(m, gauge.Metrics)
	}
	return m
}
func (s *MemoryService) Gauges() MetricStorage[*metrics.Gauge] {
	return s.gauges
}

func (s *MemoryService) Counters() MetricStorage[*metrics.Counter] {
	return s.counters
}

func (s *MemoryService) ReadBackup() error {
	return nil
}

func (s *MemoryService) WriteBackup() error {
	return nil
}

func (s *BackupMemoryService) GetAllMetrics() []metrics.Metrics {
	var m []metrics.Metrics
	for _, counter := range s.backupStorage.Counters().All() {
		m = append(m, counter.Metrics)
	}
	for _, gauge := range s.backupStorage.Gauges().All() {
		m = append(m, gauge.Metrics)
	}
	return m
}

func (s *BackupMemoryService) Gauges() MetricStorage[*metrics.Gauge] {
	return s.backupStorage.Gauges()
}

func (s *BackupMemoryService) Counters() MetricStorage[*metrics.Counter] {
	return s.backupStorage.Counters()
}

func (s *BackupMemoryService) BackupGauges() BackupMetricStorage[*metrics.Gauge] {
	return s.backupStorage.Gauges()
}

func (s *BackupMemoryService) BackupCounters() BackupMetricStorage[*metrics.Counter] {
	return s.backupStorage.Counters()
}

func (s *BackupMemoryService) WriteBackup() error {
	if err := s.backupStorage.Reader.Reset(); err != nil {
		return err
	}
	metricsToWrite := s.GetAllMetrics()
	if err := s.backupStorage.Writer.file.Truncate(0); err != nil {
		return err
	}
	if _, err := s.backupStorage.Writer.file.Seek(0, 0); err != nil {
		return err
	}
	s.backupStorage.Writer.writer.Reset(s.backupStorage.Writer.file)

	return s.backupStorage.MetricsHandler.WriteBatch(metricsToWrite)
}

func (s *BackupMemoryService) ReadBackup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.backupStorage.Reader.Reset(); err != nil {
		return err
	}

	metricsData, err := s.backupStorage.MetricsHandler.Read()
	if err != nil {
		return err
	}

	for _, metric := range metricsData {
		switch metric.MType {
		case common.Counter:
			counter := &metrics.Counter{Metrics: metric}
			if err := s.backupStorage.Counters().Set(metric.ID, counter); err != nil {
				return err
			}
		case common.Gauge:
			gauge := &metrics.Gauge{Metrics: metric}
			if err := s.backupStorage.Gauges().Set(metric.ID, gauge); err != nil {
				return err
			}
		default:
			continue
		}
	}

	return nil
}
