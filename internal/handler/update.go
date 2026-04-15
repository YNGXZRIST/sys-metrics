package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	collector "sys-metrics/internal/agent"
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/observer"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"sys-metrics/internal/service/responsewriter"
	"time"

	"go.uber.org/zap"
)

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	id := r.PathValue("name")
	value := r.PathValue("value")
	metricType = strings.ToLower(metricType)
	id = collector.GetMetricType(id)
	ctx := r.Context()
	err := serviceMetrics.Update(ctx, metricType, id, value)
	if err != nil {
		h.Logger.Warn("UpdateHandler got error", zap.Error(err))
		responsewriter.WriteBadRequest(w)
		return
	}
	obs, err := h.GetObserverByType(ObserverAudit)
	if err != nil {
		h.Logger.Warn("UpdateHandlerJSON got error", zap.Error(err))
	} else {
		event := observer.MetricsEvent{
			TS:      time.Now().UTC().Unix(),
			IP:      h.GetIPFromRequest(r),
			Metrics: []string{id},
		}
		obs.Notify(event)
	}
	responsewriter.WriteSuccess(w)
}

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
	obs, err := h.GetObserverByType(ObserverAudit)
	if err != nil {
		h.Logger.Warn("UpdateHandlerJSON got error", zap.Error(err))
	} else {
		event := observer.MetricsEvent{
			TS:      time.Now().UTC().Unix(),
			IP:      h.GetIPFromRequest(r),
			Metrics: []string{req.ID},
		}
		obs.Notify(event)
	}

	responsewriter.WriteSuccessStatus(w)
	enc := json.NewEncoder(w)
	if err := enc.Encode(metric); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
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
	err := serviceMetrics.BatchUpdateMetrics(ctx, req)
	if err != nil {
		h.Logger.Warn("UpdatesMetricsHandlerJSON got error", zap.Error(err))
		responsewriter.WriteServerError(w)
		return
	}
	obs, err := h.GetObserverByType(ObserverAudit)
	if err != nil {
		h.Logger.Warn("UpdatesMetricsHandlerJSON got error", zap.Error(err))
	} else {
		mNames := make([]string, 0, len(req))
		for _, m := range req {
			mNames = append(mNames, m.ID)
		}
		event := observer.MetricsEvent{
			TS:      time.Now().UTC().Unix(),
			IP:      h.GetIPFromRequest(r),
			Metrics: mNames,
		}
		obs.Notify(event)
	}
	responsewriter.WriteSuccessStatus(w)

}
