package middleware

import (
	"bytes"
	"io"
	"net/http"
	"sys-metrics/internal/common"
	"sys-metrics/internal/secure"
	"sys-metrics/internal/service/responsewriter"

	"go.uber.org/zap"
)

// SecureMiddleware reads the request body. When the decryptor is enabled and the request carries
// EncryptHeader=RSA (see common.EncryptHeader / common.RSA, same values the agent sets when encrypting),
// the body is decrypted; otherwise the body is passed through unchanged. If decryptor is nil or disabled,
// no decryption is attempted. Body read or decrypt errors yield HTTP 400. After successful decryption
// the encrypt header is removed before the next handler runs.
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
			if r.Header.Get(common.EncryptHeader) != common.RSA {
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
			r.Header.Del(common.EncryptHeader)
			r.Body = io.NopCloser(bytes.NewBuffer(decBody))
			next.ServeHTTP(w, r)
		})
	}
}
