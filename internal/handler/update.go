package handler

import (
	"encoding/json"
	"fmt"
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

type RequestBody struct {
	Id    string `json:"id"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func UpdateHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var req RequestBody
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := serviceMetrics.Update(req.Type, req.Id, req.Value)
	if err != nil {
		responsewriter.WriteBadRequest(w)
		return
	}
	metric, err := getMetricFromStorage(req.Type, req.Id)
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
