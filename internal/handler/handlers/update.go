package handlers

import (
	"net/http"
	serviceMetrics "sys-metrics/internal/service/metrics/methods"
	"sys-metrics/internal/service/responsewriter"
)

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	name := r.PathValue("name")
	value := r.PathValue("value")
	err := serviceMetrics.Update(metricType, name, value)
	if err != nil {
		responsewriter.WriteBadRequest(w)
		return
	}
	responsewriter.WriteSuccess(w)
}
