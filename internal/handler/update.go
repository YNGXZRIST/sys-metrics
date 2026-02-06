package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	collector "sys-metrics/internal/agent"
	"sys-metrics/internal/common"
	"sys-metrics/internal/context"
	models "sys-metrics/internal/model/metrics"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"sys-metrics/internal/service/responsewriter"
)

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	id := r.PathValue("name")
	value := r.PathValue("value")
	metricType = strings.ToLower(metricType)
	id = collector.GetMetricType(id)
	err := serviceMetrics.Update(metricType, id, value)
	if err != nil {
		responsewriter.WriteBadRequest(w)
		return
	}
	responsewriter.WriteSuccess(w)
}

func UpdateHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var req models.Metrics
	logger := context.LoggerFromContext(r.Context())
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Info("UpdateHandlerJSON got decode error: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Info(fmt.Sprintf("UpdateHandlerJSON got request: %+v", req))
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
	err := serviceMetrics.Update(req.MType, req.ID, valueStr)
	if err != nil {
		responsewriter.WriteBadRequest(w)
		return
	}
	metric, err := getMetricFromStorage(req.MType, req.ID)
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
