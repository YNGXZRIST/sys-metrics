package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	collector "sys-metrics/internal/agent"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/errors/timeerrors"
	models "sys-metrics/internal/model/metrics"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"sys-metrics/internal/service/responsewriter"
	"sys-metrics/pkg/storage"

	"go.uber.org/zap"
)

// ValueHandler handles GET /value/{type}/{name} and writes the value as plain text.
func (h *Handler) ValueHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricType := r.PathValue("type")
	name := r.PathValue("name")
	metricType = strings.ToLower(metricType)
	name = collector.GetMetricType(name)
	v, err := h.MetricService.GetMetric(ctx, metricType, name)
	if err != nil {
		h.Logger.Error("Failed to get metric", zap.Error(labelerrors.NewLabelError("VALUE", timeerrors.NewTimeError(err))))
		writeServerValueError(w, err)
		return
	}
	responsewriter.WriteSuccessStatus(w)
	writtenValue, err := h.MetricService.MetricValue(v)
	if err != nil {
		responsewriter.WriteBadRequest(w)
		return
	}
	_, err = w.Write([]byte(writtenValue))
	if err != nil {
		h.Logger.Error("Failed to write response", zap.Error(labelerrors.NewLabelError("VALUE", timeerrors.NewTimeError(err))))
		return
	}
}

// ValueHandlerJSON handles POST /value with a JSON body and returns the metric as JSON.
func (h *Handler) ValueHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var req models.Metrics
	dec := json.NewDecoder(r.Body)
	ctx := r.Context()
	if err := dec.Decode(&req); err != nil || req.ID == "" || req.MType == "" {
		responsewriter.WriteNotFound(w)
		return
	}
	v, err := h.MetricService.GetMetric(ctx, req.MType, req.ID)
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

func writeServerValueError(w http.ResponseWriter, err error) {

	if errors.Is(err, storage.ErrNotFound) {
		responsewriter.WriteNotFound(w)
		return
	}
	if errors.Is(err, serviceMetrics.ErrUnknownMetricType) {
		responsewriter.WriteBadRequest(w)
		return
	}
	responsewriter.WriteServerError(w)
}
