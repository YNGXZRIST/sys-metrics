package agent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/secure"
	"testing"

	"go.uber.org/zap"
)

func writePubPEM(t *testing.T) string {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "public.pem")
	block := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&priv.PublicKey)}
	if err := os.WriteFile(p, pem.EncodeToMemory(block), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestReporter_withEncryptor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`OK`))
	}))
	t.Cleanup(srv.Close)

	pubPath := writePubPEM(t)
	enc, err := secure.NewRequestEncryptor(pubPath)
	if err != nil {
		t.Fatal(err)
	}
	rep := NewReporter(srv.URL, zap.NewNop(), nil, enc)
	col := NewCollector(context.Background(), 1)
	g := models.NewGauge("e")
	g.SetValue(1)
	col.Gauges["e"] = g
	if err := rep.Send(col); err != nil {
		t.Fatal(err)
	}
}
