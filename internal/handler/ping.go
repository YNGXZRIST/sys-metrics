package handler

import (
	"net/http"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/errors/pgerrors"
	"sys-metrics/internal/service/responsewriter"

	"go.uber.org/zap"
)

func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.Logger
	if h.Conn == nil {
		responsewriter.WriteSuccess(w)
		return
	}
	err := h.Conn.PingContext(ctx)
	if err != nil {
		log.Error("failed to ping db connection from context", zap.Error(labelerrors.NewLabelError("DB", pgerrors.NewPgError(err))))
		responsewriter.WriteInternalServerError(w)
		return
	}
	responsewriter.WriteSuccess(w)

}
