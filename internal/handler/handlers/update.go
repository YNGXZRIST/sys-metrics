package handlers

import (
	"fmt"
	"net/http"
	serviceMetrics "sys-metrics/internal/service/metrics/update"
)

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	name := r.PathValue("name")
	value := r.PathValue("value")
	err := serviceMetrics.Update(metricType, name, value)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w)

}
func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_, err = w.Write([]byte("400 Bad Request: " + err.Error()))
	if err != nil {
		return
	}
}
func writeSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("200 OK"))
	if err != nil {
		fmt.Println(err)
	}
}
