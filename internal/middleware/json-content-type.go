package middleware

import (
	"net/http"
	"sys-metrics/internal/common"
	"sys-metrics/internal/service/responsewriter"
)

func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := r.Header.Get(common.ContentTypeHeader)
		if res != common.ApplicationJSON {
			responsewriter.WriteUnsupportedMediaType(w)
			return
		}
		w.Header().Set(common.ContentTypeHeader, common.ApplicationJSON)
		next.ServeHTTP(w, r)

	})
}
