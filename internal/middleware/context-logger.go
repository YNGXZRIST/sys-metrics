package middleware

import (
	"context"
	"net/http"
	"sys-metrics/internal/common"

	"go.uber.org/zap"
)

// WithLoggerContext stores zap.Logger in the request context under common.ContextLoggerKey.
func WithLoggerContext(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), common.ContextLoggerKey, logger)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
