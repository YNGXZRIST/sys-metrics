package handler

import (
	"html/template"
	"io/fs"
	"net/http"
	"sys-metrics/internal"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	svm "sys-metrics/internal/repository"
	"sys-metrics/internal/service/responsewriter"
)

type PageData struct {
	Gauge   map[string]*metrics.Gauge
	Counter map[string]*metrics.Counter
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	sub, err := fs.Sub(internal.StaticFS, "static")
	if err != nil {
		responsewriter.WriteServerError(w)
		return
	}
	tmpl, err := template.ParseFS(sub, "index.html")
	if err != nil {
		responsewriter.WriteServerError(w)
		return
	}
	data := PageData{
		Gauge:   svm.Gauges().All(),
		Counter: svm.Counters().All(),
	}
	w.Header().Set(common.ContentTypeHeader, common.TextHTMLUTF8)
	if err := tmpl.Execute(w, data); err != nil {
		responsewriter.WriteServerError(w)
		return
	}
}
