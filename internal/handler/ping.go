package handler

import (
	"net/http"
	"sys-metrics/internal/service/responsewriter"

	"go.uber.org/zap"
)
import ctxUtil "sys-metrics/internal/context"

func PingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := ctxUtil.LoggerFromContext(r.Context())
	conn, err := ctxUtil.DBFromContext(r.Context())
	if err != nil {
		log.Error("failed to get db connection from context", zap.Error(err))
		responsewriter.WriteInternalServerError(w)
		return
	}
	err = conn.PingContext(ctx)
	if err != nil {
		log.Error("failed to ping db connection from context", zap.Error(err))
		responsewriter.WriteInternalServerError(w)
		return
	}
	responsewriter.WriteSuccess(w)

}
