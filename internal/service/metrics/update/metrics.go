package update

import (
	"strconv"
	"sys-metrics/internal/repository/metrics/storagemetrics"
	"sys-metrics/pkg/memstorage"
	"sys-metrics/pkg/metrics/counter"
	"sys-metrics/pkg/metrics/gauge"
)

func Update(metricType, name, value string) error {
	storage, err := storagemetrics.StorageFactory(metricType)
	if err != nil {
		return err
	}
	switch s := storage.(type) {
	case *memstorage.MemStorage[string, *counter.Metric]:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return updateOrCreate(s, name, func() *counter.Metric {
			return counter.NewCounter(name)
		}, parsed)

	case *memstorage.MemStorage[string, *gauge.Metric]:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		return updateOrCreate(s, name, func() *gauge.Metric {
			return gauge.NewGauge(name)
		}, parsed)
	}
	return nil
}

func updateOrCreate[T interface{ SetValue(v V) }, V any](
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
