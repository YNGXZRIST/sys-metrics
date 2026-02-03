package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	collector "sys-metrics/internal/agent"
	"sys-metrics/internal/backup"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/logger"
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
	backupConfig := getBackupConfigFromContext(r)
	if backupConfig != nil && backupConfig.IsSyncBackup() {
		metric, err := getMetricFromStorage(metricType, id)
		if err != nil {
			responsewriter.WriteBadRequest(w)
			return
		}
		err = backupConfig.UpsertMetricToBackup(&metric)
		if err != nil {
			responsewriter.WriteBadRequest(w)
			return
		}
	}
	responsewriter.WriteSuccess(w)
}

func UpdateHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var req models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Info("UpdateHandlerJSON got decode error: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Log.Info(fmt.Sprintf("UpdateHandlerJSON got request: %+v", req))
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
	backupConfig := getBackupConfigFromContext(r)
	if backupConfig != nil && backupConfig.IsSyncBackup() {
		err = backupConfig.UpsertMetricToBackup(&metric)
		if err != nil {
			responsewriter.WriteBadRequest(w)
			return
		}
	}
	responsewriter.WriteSuccessStatus(w)
	enc := json.NewEncoder(w)
	if err := enc.Encode(metric); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
func getBackupConfigFromContext(r *http.Request) *backup.BackupConfig {
	if cfg, ok := r.Context().Value("config").(*server.Config); ok && cfg != nil {
		return cfg.BackupConfig
	}
	return nil
}
