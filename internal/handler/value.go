package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	collector "sys-metrics/internal/agent"
	"sys-metrics/internal/common"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/errors/timeerrors"
	models "sys-metrics/internal/model/metrics"
	svm "sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/service/responsewriter"
	"sys-metrics/pkg/storage"

	"go.uber.org/zap"
)

var (
	// ErrUnknownMetricType is returned when the request uses an unsupported metric type.
	ErrUnknownMetricType = fmt.Errorf("unknown metric type")
)

// ValueHandler handles GET /value/{type}/{name} and writes the value as plain text.
func (h *Handler) ValueHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricType := r.PathValue("type")
	name := r.PathValue("name")
	metricType = strings.ToLower(metricType)
	name = collector.GetMetricType(name)
	v, err := getMetricFromStorage(ctx, metricType, name)
	if err != nil {
		h.Logger.Error("Failed to get metric", zap.Error(labelerrors.NewLabelError("VALUE", timeerrors.NewTimeError(err))))
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
	v, err := getMetricFromStorage(ctx, req.MType, req.ID)
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

func getMetricFromStorage(ctx context.Context, metricType, name string) (models.Metrics, error) {
	switch metricType {
	case common.Counter:
		v, err := svm.Counters().Get(ctx, name)
		if err != nil {
			return models.Metrics{}, labelerrors.NewLabelError("COUNTER", storage.ErrNotFound)
		}
		return v.Metrics, nil
	case common.Gauge:
		v, err := svm.Gauges().Get(ctx, name)
		if err != nil {
			return models.Metrics{}, labelerrors.NewLabelError("GAUGE", storage.ErrNotFound)
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
