package middleware

import (
	"bytes"
	"io"
	"net/http"
	"sys-metrics/pkg/pool"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter // embeds the underlying ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithRequestLogger logs method, URI, status, duration, and for POST requests the body.
func WithRequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	bufPool := pool.New(func() *bytes.Buffer { return new(bytes.Buffer) })
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sugar := logger.Sugar()
			var buf *bytes.Buffer
			if r.Method == http.MethodPost {
				buf = bufPool.Get()
				tee := io.TeeReader(r.Body, buf)
				body, err := io.ReadAll(tee)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				sugar.Infoln(
					"uri", r.RequestURI,
					"method", r.Method,
					"request body", string(body),
				)
				r.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
				defer bufPool.Put(buf)
			}
			start := time.Now()
			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			h.ServeHTTP(&lw, r)
			duration := time.Since(start)
			sugar.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"status", responseData.status,
				"duration", duration,
				"size", responseData.size,
			)

		})
	}
}
