package middleware

import (
	"context"
	"net/http"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
)

// WithDBContext adds *db.DB to the request context (common.ContextDBKey) when conn is non-nil.
func WithDBContext(conn *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if conn != nil {
				r = r.WithContext(context.WithValue(r.Context(), common.ContextDBKey, conn))
			}
			next.ServeHTTP(w, r)
		})
	}
}
