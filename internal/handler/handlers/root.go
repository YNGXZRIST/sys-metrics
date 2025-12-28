package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sys-metrics/internal/model/metrics"
	svm "sys-metrics/internal/service/metrics"
	"sys-metrics/internal/service/responsewriter"
)

type PageData struct {
	Gauge   map[string]*metrics.Gauge
	Counter map[string]*metrics.Counter
}

func Index(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs("internal/views/index.html")
	if err != nil {
		responsewriter.WriteServerError(w)
		return
	}
	fmt.Println(path)
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		responsewriter.WriteServerError(w)
		return
	}
	data := PageData{
		Gauge:   svm.Gauges().All(),
		Counter: svm.Counters().All(),
	}

	_ = tmpl.Execute(w, data)
	//if err != nil {
	//	responsewriter.WriteServerError(w)
	//}
}
