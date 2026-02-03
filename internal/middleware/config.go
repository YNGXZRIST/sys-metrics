package middleware

import (
	"context"
	"net/http"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/server"
)

func WithConfig(cfg *server.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), common.ConfigContextKey, cfg)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
