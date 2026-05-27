package middleware

import (
	"net"
	"net/http"
	"sys-metrics/internal/common"
)

func TrustedSubnet(ipNet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ipNet == nil {
				next.ServeHTTP(w, r)
				return
			}
			realIP, ok := getHeaderXRealIP(r)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			clientIP := net.ParseIP(realIP)
			if clientIP == nil || ipNet.Contains(clientIP) {
				next.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusForbidden)
		})
	}
}

func getHeaderXRealIP(r *http.Request) (string, bool) {
	realIP := r.Header.Get(common.HeaderXRealIP)
	return realIP, realIP != ""
}
