package sender

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/common"
	"sys-metrics/internal/grpchandler"
	models "sys-metrics/internal/model/metrics"
	pb "sys-metrics/internal/proto"
	"sys-metrics/internal/repository/memory"
	repoMetrics "sys-metrics/internal/repository/metrics"
	"sys-metrics/internal/secure"
	serviceMetrics "sys-metrics/internal/service/metrics"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestGrpcSender_SendBatch_serverUnavailable(t *testing.T) {
	s, err := newGrpcSender(grpcSenderConfig{
		senderDeps: senderDeps{logger: testLogger()},
		endpoint:   "127.0.0.1:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	g := models.NewGauge("X")
	g.SetValue(1)
	err = s.SendBatch(context.Background(), []*models.Metrics{&g.Metrics})
	if err == nil {
		t.Fatal("expected send error")
	}
}

func testLogger() *zap.Logger {
	return zap.NewNop()
}

func TestNewMetricsSender_http(t *testing.T) {
	s, err := NewMetricsSender(Config{
		Transport: common.ReportTransportHTTP,
		ServerURL: "http://127.0.0.1:8080",
		Logger:    testLogger(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewMetricsSender_grpc(t *testing.T) {
	addr, stop := startTestGRPCServer(t)
	defer stop()

	s, err := NewMetricsSender(Config{
		Transport: common.ReportTransportGRPC,
		Endpoint:  addr,
		Logger:    testLogger(),
		LocalIPv4: "127.0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
}

func TestHttpSender_SendBatch(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != PathUpdates {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get(common.HeaderXRealIP) != "10.0.0.2" {
			t.Fatalf("X-Real-IP = %q", r.Header.Get(common.HeaderXRealIP))
		}
		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	key := "secret"
	s := newHttpSender(httpSenderConfig{
		senderDeps: senderDeps{
			logger:    testLogger(),
			localIPv4: "10.0.0.2",
		},
		authenticator: authenticate.NewSha256(&key),
		serverURL:     srv.URL,
	})

	gauge := models.NewGauge("Alloc")
	gauge.SetValue(1.5)
	counter := models.NewCounter("PollCount")
	counter.SetValue(3)
	err := s.SendBatch(context.Background(), []*models.Metrics{
		&gauge.Metrics,
		&counter.Metrics,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(gotBody) == 0 {
		t.Fatal("expected request body")
	}
}

func TestHttpSender_SendBatch_empty(t *testing.T) {
	s := newHttpSender(httpSenderConfig{
		senderDeps: senderDeps{logger: testLogger()},
		serverURL:  "http://127.0.0.1:1",
	})
	if err := s.SendBatch(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestHttpSender_sendMetricToServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != PathUpdate {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	s := newHttpSender(httpSenderConfig{
		senderDeps: senderDeps{logger: testLogger()},
		serverURL:  srv.URL,
	})
	gauge := models.NewGauge("HeapAlloc")
	gauge.SetValue(42)
	if err := s.sendMetricToServer(context.Background(), gauge.Metrics); err != nil {
		t.Fatal(err)
	}
}

func TestHttpSender_withEnabledEncryptor(t *testing.T) {
	pubPath := writeTestPublicKey(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(common.EncryptHeader) != common.RSA {
			t.Fatalf("encrypt header = %q", r.Header.Get(common.EncryptHeader))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	enc, err := secure.NewRequestEncryptor(pubPath)
	if err != nil {
		t.Fatal(err)
	}

	s := newHttpSender(httpSenderConfig{
		senderDeps:       senderDeps{logger: testLogger()},
		requestEncryptor: enc,
		serverURL:        srv.URL,
	})
	g := models.NewGauge("HeapSys")
	g.SetValue(1)
	if err := s.SendBatch(context.Background(), []*models.Metrics{&g.Metrics}); err != nil {
		t.Fatal(err)
	}
}

func writeTestPublicKey(t *testing.T) string {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	pubPath := filepath.Join(dir, "public.pem")
	pubBlock := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&priv.PublicKey)}
	if err := os.WriteFile(pubPath, pem.EncodeToMemory(pubBlock), 0o600); err != nil {
		t.Fatal(err)
	}
	return pubPath
}

func TestHttpSender_withDisabledEncryptor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	enc, err := secure.NewRequestEncryptor("")
	if err != nil {
		t.Fatal(err)
	}

	s := newHttpSender(httpSenderConfig{
		senderDeps:       senderDeps{logger: testLogger()},
		requestEncryptor: enc,
		serverURL:        srv.URL,
	})
	sys := models.NewGauge("Sys")
	sys.SetValue(0)
	if err := s.SendBatch(context.Background(), []*models.Metrics{&sys.Metrics}); err != nil {
		t.Fatal(err)
	}
}

func TestGrpcSender_SendBatch(t *testing.T) {
	addr, stop := startTestGRPCServer(t)
	defer stop()

	s, err := newGrpcSender(grpcSenderConfig{
		senderDeps: senderDeps{
			logger:    testLogger(),
			localIPv4: "127.0.0.1",
		},
		endpoint: addr,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	gauge := models.NewGauge("HeapInuse")
	gauge.SetValue(2.5)
	err = s.SendBatch(context.Background(), []*models.Metrics{&gauge.Metrics})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGrpcSender_SendBatch_empty(t *testing.T) {
	s := &grpcSender{grpcSenderConfig: grpcSenderConfig{senderDeps: senderDeps{logger: testLogger()}}}
	if err := s.SendBatch(context.Background(), nil); err == nil {
		t.Fatal("expected error for empty batch")
	}
}

func TestGrpcSender_getProtoMetrics(t *testing.T) {
	s := &grpcSender{}
	g1 := models.NewGauge("G1")
	g1.SetValue(1.1)
	c1 := models.NewCounter("C1")
	c1.SetValue(5)
	got := s.getProtoMetrics([]*models.Metrics{
		nil,
		&g1.Metrics,
		&c1.Metrics,
	})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].GetType() != pb.Metric_GAUGE || got[0].GetValue() != 1.1 {
		t.Fatalf("gauge: %+v", got[0])
	}
	if got[1].GetType() != pb.Metric_COUNTER || got[1].GetDelta() != 5 {
		t.Fatalf("counter: %+v", got[1])
	}
}

func TestGrpcSender_Close_nilConn(t *testing.T) {
	s := &grpcSender{}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func startTestGRPCServer(t *testing.T) (addr string, stop func()) {
	t.Helper()

	repoMetrics.Init(memory.NewService())
	svc := serviceMetrics.NewService(nil)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterMetricsServer(grpcSrv, grpchandler.NewMetricServer(svc, testLogger()))

	go grpcSrv.Serve(lis)

	return lis.Addr().String(), func() {
		grpcSrv.Stop()
		_ = lis.Close()
	}
}
