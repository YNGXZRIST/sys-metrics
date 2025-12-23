package handlers

import (
	"fmt"
	"net/http"
	serviceMetrics "sys-metrics/internal/service/metrics/methods"
)

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	name := r.PathValue("name")
	value := r.PathValue("value")
	err := serviceMetrics.Update(metricType, name, value)
	if err != nil {
		write404(w)
		return
	}
	writeSuccess(w)

}
func write404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}
func writeSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("200 OK"))
	if err != nil {
		fmt.Println(err)
	}
}
