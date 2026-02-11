package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	collector "sys-metrics/internal/agent"
	"sys-metrics/internal/common"
	"sys-metrics/internal/context"
	models "sys-metrics/internal/model/metrics"
	svm "sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/service/responsewriter"
	"sys-metrics/pkg/storage"

	"go.uber.org/zap"
)

var (
	ErrUnknownMetricType = fmt.Errorf("unknown metric type")
)

func ValueHandler(w http.ResponseWriter, r *http.Request) {
	logger := context.LoggerFromContext(r.Context())
	metricType := r.PathValue("type")
	name := r.PathValue("name")
	metricType = strings.ToLower(metricType)
	name = collector.GetMetricType(name)
	v, err := getMetricFromStorage(metricType, name)
	if err != nil {
		writeServerValueError(w, err)
		return
	}
	responsewriter.WriteSuccessStatus(w)
	var writtenValue string
	switch v.MType {
	case common.Counter:
		writtenValue = strconv.FormatInt(*v.Delta, 10)
	case common.Gauge:
		writtenValue = strconv.FormatFloat(*v.Value, 'f', -1, 64)
	default:
		responsewriter.WriteBadRequest(w)
		return

	}
	_, err = w.Write([]byte(writtenValue))
	if err != nil {
		logger.Warn("Failed to write response", zap.Error(err))
		return
	}
}

func ValueHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var req models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil || req.ID == "" || req.MType == "" {
		responsewriter.WriteNotFound(w)
		return
	}
	v, err := getMetricFromStorage(req.MType, req.ID)
	if err != nil {
		writeServerValueError(w, err)
		return
	}
	responsewriter.WriteSuccessStatus(w)
	enc := json.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func getMetricFromStorage(metricType, name string) (models.Metrics, error) {
	switch metricType {
	case common.Counter:
		v, err := svm.Counters().Get(name)
		if err != nil {
			return models.Metrics{}, storage.ErrNotFound
		}
		return v.Metrics, nil
	case common.Gauge:
		v, err := svm.Gauges().Get(name)
		if err != nil {
			return models.Metrics{}, storage.ErrNotFound
		}
		return v.Metrics, nil
	default:
		return models.Metrics{}, ErrUnknownMetricType
	}
}
func writeServerValueError(w http.ResponseWriter, err error) {

	if errors.Is(err, storage.ErrNotFound) {
		responsewriter.WriteNotFound(w)
		return
	}
	if errors.Is(err, ErrUnknownMetricType) {
		responsewriter.WriteBadRequest(w)
		return
	}
	responsewriter.WriteServerError(w)
}
