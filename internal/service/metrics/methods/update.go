package methods

import (
	"strconv"
	"sys-metrics/internal/model/metrics"
	s "sys-metrics/internal/service/metrics"
	"sys-metrics/pkg/memstorage"
)

func Update(metricType, name, value string) error {
	storage, err := s.StorageFactory(metricType)
	if err != nil {
		return err
	}
	switch s := storage.(type) {
	case *memstorage.MemStorage[string, *metrics.Counter]:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return updateOrCreate(s, name, func() *metrics.Counter {
			return metrics.NewCounter(name)
		}, parsed)

	case *memstorage.MemStorage[string, *metrics.Gauge]:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		return updateOrCreate(s, name, func() *metrics.Gauge {
			return metrics.NewGauge(name)
		}, parsed)
	}
	return nil
}

func updateOrCreate[T interface{ SetValue(v V) V }, V any](
	s *memstorage.MemStorage[string, T],
	name string,
	factory func() T,
	value V,
) error {
	metric, err := s.Get(name)
	if err != nil {
		metric = factory()
	}
	metric.SetValue(value)
	return s.Set(name, metric)
}
