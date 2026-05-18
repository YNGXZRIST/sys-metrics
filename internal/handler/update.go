package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	collector "sys-metrics/internal/agent"
	models "sys-metrics/internal/model/metrics"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"sys-metrics/internal/service/responsewriter"

	"go.uber.org/zap"
)

// UpdateHandler handles POST /update/{type}/{name}/{value} (plain-text update of one metric).
func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	id := r.PathValue("name")
	value := r.PathValue("value")
	metricType = strings.ToLower(metricType)
	id = collector.GetMetricType(id)
	ctx := r.Context()
	err := h.MetricService.Update(ctx, metricType, id, value)
	if err != nil {
		h.Logger.Warn("UpdateHandler got error", zap.Error(err))
		responsewriter.WriteBadRequest(w)
		return
	}
	responsewriter.WriteSuccess(w)
}

// UpdateHandlerJSON handles POST /update with a JSON body (one metric) and returns the current state in the response.
func (h *Handler) UpdateHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var req models.Metrics
	ctx := r.Context()
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.Logger.Warn("UpdateHandlerJSON got error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.Logger.Info("UpdateHandlerJSON", zap.Any("req", req))
	metric, err := h.MetricService.UpdateMetric(ctx, req)
	if err != nil {
		if errors.Is(err, serviceMetrics.ErrMetricRead) {
			responsewriter.WriteServerError(w)
			return
		}
		responsewriter.WriteBadRequest(w)
		return
	}

	responsewriter.WriteSuccessStatus(w)
	enc := json.NewEncoder(w)
	if err := enc.Encode(metric); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

// UpdatesMetricsHandlerJSON handles POST /updates with a JSON array of metrics (batch update).
func (h *Handler) UpdatesMetricsHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var req []models.Metrics
	ctx := r.Context()
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.Logger.Warn("UpdatesMetricsHandlerJSON got error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.Logger.Info("UpdatesMetricsHandlerJSON", zap.Any("req", req))
	err := h.MetricService.BatchUpdateMetrics(ctx, req)
	if err != nil {
		h.Logger.Warn("UpdatesMetricsHandlerJSON got error", zap.Error(err))
		responsewriter.WriteServerError(w)
		return
	}
	responsewriter.WriteSuccessStatus(w)

}
