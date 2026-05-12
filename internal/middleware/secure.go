package middleware

import (
	"bytes"
	"io"
	"net/http"
	"sys-metrics/internal/secure"
	"sys-metrics/internal/service/responsewriter"

	"go.uber.org/zap"
)

// SecureMiddleware reads the entire request body, optionally decrypts it when decryptor is enabled,
// then restores r.Body for the next handler. If decryptor is nil or IsEnabled is false, the body is
// passed through unchanged. Decryption or body read failures yield HTTP 400.
func SecureMiddleware(logger *zap.Logger, d *secure.RequestDecryptor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("failed to read body", zap.Error(err))
				responsewriter.WriteBadRequest(w)
				return
			}
			if d == nil || !d.IsEnabled {
				r.Body = io.NopCloser(bytes.NewBuffer(body))
				next.ServeHTTP(w, r)
				return
			}
			decBody, errD := d.Decrypt(body)
			if errD != nil {
				logger.Error("failed to decrypt body", zap.Error(err))
				responsewriter.WriteBadRequest(w)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(decBody))
			next.ServeHTTP(w, r)
		})
	}
}
