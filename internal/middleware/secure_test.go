package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sys-metrics/internal/common"
	"sys-metrics/internal/secure"
	"testing"

	"go.uber.org/zap"
)

func writeTestRSAPEM(t *testing.T) (pubPath, privPath string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	privPath = filepath.Join(dir, "private.pem")
	pubPath = filepath.Join(dir, "public.pem")
	privBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}
	if err := os.WriteFile(privPath, pem.EncodeToMemory(privBlock), 0o600); err != nil {
		t.Fatal(err)
	}
	pubBlock := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&priv.PublicKey)}
	if err := os.WriteFile(pubPath, pem.EncodeToMemory(pubBlock), 0o600); err != nil {
		t.Fatal(err)
	}
	return pubPath, privPath
}

func TestSecureMiddleware_nilDecryptor(t *testing.T) {
	var nextBody []byte
	h := SecureMiddleware(zap.NewNop(), nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusTeapot)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("hello"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d", rec.Code)
	}
	if string(nextBody) != "hello" {
		t.Fatalf("body = %q", nextBody)
	}
}

func TestSecureMiddleware_decryptorDisabled(t *testing.T) {
	dec, err := secure.NewRequestDecryptor("")
	if err != nil {
		t.Fatal(err)
	}
	h := SecureMiddleware(zap.NewNop(), dec)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if string(b) != "plain" {
			t.Errorf("got %q", b)
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("plain"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSecureMiddleware_decryptFails(t *testing.T) {
	_, privPath := writeTestRSAPEM(t)
	dec, err := secure.NewRequestDecryptor(privPath)
	if err != nil {
		t.Fatal(err)
	}
	h := SecureMiddleware(zap.NewNop(), dec)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not run")
	}))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not-encrypted"))
	req.Header.Set(common.EncryptHeader, common.RSA)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestSecureMiddleware_decryptOK(t *testing.T) {
	pubPath, privPath := writeTestRSAPEM(t)
	enc, err := secure.NewRequestEncryptor(pubPath)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := secure.NewRequestDecryptor(privPath)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"ok":true}`)
	blob, err := enc.Encrypt(payload)
	if err != nil {
		t.Fatal(err)
	}
	var got []byte
	h := SecureMiddleware(zap.NewNop(), dec)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = io.ReadAll(r.Body)
		if r.Header.Get(common.EncryptHeader) != "" {
			t.Fatal("encrypt header should be removed after decrypt")
		}
		w.WriteHeader(http.StatusCreated)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(blob))
	req.Header.Set(common.EncryptHeader, common.RSA)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d", rec.Code)
	}
	if string(got) != string(payload) {
		t.Fatalf("body = %q want %q", got, payload)
	}
}

func TestSecureMiddleware_noEncryptHeader_passthrough(t *testing.T) {
	_, privPath := writeTestRSAPEM(t)
	dec, err := secure.NewRequestDecryptor(privPath)
	if err != nil {
		t.Fatal(err)
	}
	var got []byte
	h := SecureMiddleware(zap.NewNop(), dec)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	plain := []byte(`{"x":1}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(plain))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if string(got) != string(plain) {
		t.Fatalf("body = %q", got)
	}
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestSecureMiddleware_bodyReadError(t *testing.T) {
	h := SecureMiddleware(zap.NewNop(), nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not run")
	}))
	req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(errReader{}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}
