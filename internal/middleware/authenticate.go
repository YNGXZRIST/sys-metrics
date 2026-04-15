package middleware

import (
	"bytes"
	"io"
	"net/http"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/service/responsewriter"

	"go.uber.org/zap"
)

type signingResponseWriter struct {
	http.ResponseWriter
	authenticator authenticate.Authenticator
	buf           *bytes.Buffer
	statusCode    int
}

func newSigningResponseWriter(w http.ResponseWriter, authenticator authenticate.Authenticator) *signingResponseWriter {
	return &signingResponseWriter{
		ResponseWriter: w,
		authenticator:  authenticator,
		buf:            &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

func (s *signingResponseWriter) Header() http.Header {
	return s.ResponseWriter.Header()
}

func (s *signingResponseWriter) Write(b []byte) (int, error) {
	return s.buf.Write(b)
}

func (s *signingResponseWriter) WriteHeader(statusCode int) {
	s.statusCode = statusCode
}

func (s *signingResponseWriter) flush() error {
	body := s.buf.Bytes()
	hash := s.authenticator.SignBody(body)
	s.ResponseWriter.Header().Set(s.authenticator.GetHashHeaderKey(), hash)
	s.ResponseWriter.WriteHeader(s.statusCode)
	_, err := s.ResponseWriter.Write(body)
	return err
}

// WithAuthenticateMiddleware validates the HMAC header on incoming requests and signs responses when authenticator is set.
func WithAuthenticateMiddleware(logger *zap.Logger, authenticator authenticate.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authenticator == nil {
				next.ServeHTTP(w, r)
				return
			}
			header := r.Header.Get(authenticator.GetHashHeaderKey())
			if header != "" {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					logger.Error("failed to read body", zap.Error(err))
					responsewriter.WriteBadRequest(w)
					return
				}
				_ = r.Body.Close()
				r.Body = io.NopCloser(bytes.NewReader(body))
				valid, err := authenticator.Validate(header, body)
				if err != nil {
					logger.Error("failed to validate header", zap.String("header", authenticator.GetHashHeaderKey()), zap.Error(err))
					responsewriter.WriteBadRequest(w)
					return
				}
				if !valid {
					logger.Info("invalid header", zap.String("header", authenticator.GetHashHeaderKey()))
					responsewriter.WriteBadRequest(w)
					return
				}
			}
			signingWriter := newSigningResponseWriter(w, authenticator)
			next.ServeHTTP(signingWriter, r)
			if err := signingWriter.flush(); err != nil {
				logger.Error("failed to flush signed response", zap.Error(err))
			}
		})
	}
}
