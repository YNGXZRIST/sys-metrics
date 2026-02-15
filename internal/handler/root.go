package handler

import (
	"html/template"
	"io/fs"
	"net/http"
	"sys-metrics/internal"
	"sys-metrics/internal/common"
	ctxsrv "sys-metrics/internal/context"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/model/metrics"
	svm "sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/service/responsewriter"

	"go.uber.org/zap"
)

type PageData struct {
	Gauge   map[string]*metrics.Gauge
	Counter map[string]*metrics.Counter
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := ctxsrv.LoggerFromContext(ctx)

	sub, err := fs.Sub(internal.StaticFS, "static")
	if err != nil {
		logger.Error("failed to sub static files", zap.Error(labelerrors.NewLabelError("FS", err)))
		responsewriter.WriteServerError(w)
		return
	}
	tmpl, err := template.ParseFS(sub, "index.html")
	if err != nil {
		responsewriter.WriteServerError(w)
		return
	}
	data := PageData{
		Gauge:   svm.Gauges().All(ctx),
		Counter: svm.Counters().All(ctx),
	}
	w.Header().Set(common.ContentTypeHeader, common.TextHTMLUTF8)
	if err := tmpl.Execute(w, data); err != nil {
		logger.Error("failed to execute template", zap.Error(labelerrors.NewLabelError("TMPL", err)))
		responsewriter.WriteServerError(w)
		return
	}
}
