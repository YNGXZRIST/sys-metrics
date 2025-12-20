package handlers

import (
	"net/http"
	"strconv"
	"sys-metrics/pkg/mem_storage"
	"sys-metrics/pkg/metrics"
	"sys-metrics/pkg/metrics/counter"
	"sys-metrics/pkg/metrics/gauge"
	"sys-metrics/pkg/server"
)

type metric interface {
	*counter.Metric | *gauge.Metric
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	name := r.PathValue("name")
	value := r.PathValue("value")
	storage, err := storageFactory(metricType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err2 := w.Write([]byte("400 Bad Request: " + err.Error()))
		if err2 != nil {
			writeError(w, err2)
		}
		return
	}
	fl := false
	switch s := storage.(type) {
	case *mem_storage.MemStorage[string, *counter.Metric]:
		fl = s.Has(name)
		parsed, e := strconv.ParseInt(value, 10, 64)
		if e != nil {
			writeError(w, err)
		}
		if !fl {
			metric := counter.NewCounter(name)
			metric.SetValue(parsed)
			updateStorageMetric(w, s, name, metric)
		} else {
			metric, err := s.Get(name)
			if err != nil {
				writeError(w, err)
			}
			updateStorageMetric(w, s, name, metric)
		}
	case *mem_storage.MemStorage[string, *gauge.Metric]:
		fl = s.Has(name)
		parsed, e := strconv.ParseFloat(value, 64)
		if e != nil {
			writeError(w, err)
		}
		if !fl {
			metric := gauge.NewGauge(name)
			metric.SetValue(parsed)
			updateStorageMetric(w, s, name, metric)
		}
	}
}
func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_, err = w.Write([]byte("400 Bad Request: " + err.Error()))
	if err != nil {
		return
	}
}
func updateStorageMetric[T metric](w http.ResponseWriter, s *mem_storage.MemStorage[string, T], name string, metric T) {
	err := s.Set(name, metric)
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("200 OK"))
	if err != nil {
		return
	}
}

func storageFactory(t string) (any, error) {
	switch t {
	case "counter":
		return metrics.CounterStorage, nil
	case "gauge":
		return metrics.GaugeStorage, nil
	default:
		return nil, server.ErrUnknownMetricType
	}
}
