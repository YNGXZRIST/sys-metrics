package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	collector "sys-metrics/internal/agent"
	"sys-metrics/internal/common"
	"sys-metrics/internal/context"
	models "sys-metrics/internal/model/metrics"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"sys-metrics/internal/service/responsewriter"

	"go.uber.org/zap"
)

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	id := r.PathValue("name")
	value := r.PathValue("value")
	metricType = strings.ToLower(metricType)
	id = collector.GetMetricType(id)
	ctx := r.Context()
	err := serviceMetrics.Update(ctx, metricType, id, value)
	logger := context.LoggerFromContext(ctx)
	if err != nil {
		logger.Warn("UpdateHandler got error", zap.Error(err))
		responsewriter.WriteBadRequest(w)
		return
	}
	responsewriter.WriteSuccess(w)
}

func UpdateHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var req models.Metrics
	ctx := r.Context()
	logger := context.LoggerFromContext(ctx)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Warn("UpdateHandlerJSON got error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Info("UpdateHandlerJSON", zap.Any("req", req))
	var valueStr string
	switch req.MType {
	case common.Gauge:
		var val float64
		if req.Value == nil {
			val = 0.0
		} else {
			val = *req.Value
		}
		valueStr = strconv.FormatFloat(val, 'f', -1, 64)
	case common.Counter:
		var delta int64
		if req.Delta == nil {
			delta = 0
		} else {
			delta = *req.Delta
		}
		valueStr = strconv.FormatInt(delta, 10)
	default:
		responsewriter.WriteBadRequest(w)
		return
	}
	err := serviceMetrics.Update(ctx, req.MType, req.ID, valueStr)
	if err != nil {
		responsewriter.WriteBadRequest(w)
		return
	}
	metric, err := getMetricFromStorage(ctx, req.MType, req.ID)
	if err != nil {
		responsewriter.WriteServerError(w)
		return
	}
	responsewriter.WriteSuccessStatus(w)
	enc := json.NewEncoder(w)
	if err := enc.Encode(metric); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
