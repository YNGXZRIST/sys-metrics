package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const CtxClientIPKey contextKey = "ClientIP"

func WithClientIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}
		ctx := context.WithValue(r.Context(), CtxClientIPKey, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
