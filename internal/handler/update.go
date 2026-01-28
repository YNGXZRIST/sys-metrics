package handler

import (
	"net/http"
	"strings"
	collector "sys-metrics/internal/agent"
	serviceMetrics "sys-metrics/internal/service/metrics/methods"
	"sys-metrics/internal/service/responsewriter"
)

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	name := r.PathValue("name")
	value := r.PathValue("value")
	metricType = strings.ToLower(metricType)
	name = collector.GetMetricType(name)
	err := serviceMetrics.Update(metricType, name, value)
	if err != nil {

		responsewriter.WriteBadRequest(w)
		return
	}
	responsewriter.WriteSuccess(w)
}
