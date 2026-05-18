package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/authenticate"
	"testing"

	"go.uber.org/zap"
)

func TestWithAuthenticateMiddleware_nilAuthenticator(t *testing.T) {
	h := WithAuthenticateMiddleware(zap.NewNop(), nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("x"))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestWithAuthenticateMiddleware_signsResponse(t *testing.T) {
	key := "secret"
	auth := authenticate.NewSha256(&key)
	h := WithAuthenticateMiddleware(zap.NewNop(), auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"a":1}`))
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	hdr := rec.Header().Get(auth.GetHashHeaderKey())
	if hdr == "" {
		t.Fatal("missing signature header")
	}
}

func TestWithAuthenticateMiddleware_invalidSignature(t *testing.T) {
	key := "secret"
	auth := authenticate.NewSha256(&key)
	h := WithAuthenticateMiddleware(zap.NewNop(), auth)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next must not run")
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	req.Header.Set(auth.GetHashHeaderKey(), "deadbeef")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestWithAuthenticateMiddleware_badHexHeader(t *testing.T) {
	key := "secret"
	auth := authenticate.NewSha256(&key)
	h := WithAuthenticateMiddleware(zap.NewNop(), auth)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next must not run")
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	req.Header.Set(auth.GetHashHeaderKey(), "zz")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestWithAuthenticateMiddleware_bodyReadError(t *testing.T) {
	key := "secret"
	auth := authenticate.NewSha256(&key)
	h := WithAuthenticateMiddleware(zap.NewNop(), auth)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next must not run")
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(errReader{}))
	req.Header.Set(auth.GetHashHeaderKey(), "00")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}
