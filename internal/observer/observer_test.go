package observer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"sys-metrics/internal/common"
	"testing"
	"time"
)

func TestMetricsObserver_Register_InvalidType(t *testing.T) {
	obs, err := NewMetricsObserver(MetricObserverConfig{Mode: common.TypeModeTest, RateLimit: 1})
	if err != nil {
		t.Fatalf("NewMetricsObserver() err = %v, want nil", err)
	}

	if err := obs.Register("not-a-context"); err == nil {
		t.Fatalf("Register() err = nil, want non-nil")
	}
}

func TestMetricsObserver_RegisterAndNotify_Table(t *testing.T) {
	type testCase struct {
		name                string
		registerWith        any
		wantRegisterErr     bool
		wantWorkerPoolAfter bool
		notifyWith          any
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []testCase{
		{
			name:            "register invalid type",
			registerWith:    "not-a-context",
			wantRegisterErr: true,
		},
		{
			name:                "register success sets worker pool",
			registerWith:        ctx,
			wantRegisterErr:     false,
			wantWorkerPoolAfter: true,
		},
		{
			name:       "notify without register does nothing",
			notifyWith: MetricsEvent{Ts: 1, IP: "127.0.0.1", Metrics: []string{"m1"}},
		},
		{
			name:                "notify invalid type does nothing",
			registerWith:        ctx,
			wantWorkerPoolAfter: true,
			notifyWith:          123,
		},
		{
			name:                "notify valid event after register",
			registerWith:        ctx,
			wantWorkerPoolAfter: true,
			notifyWith:          MetricsEvent{Ts: 2, IP: "127.0.0.2", Metrics: []string{"m2"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs, err := NewMetricsObserver(MetricObserverConfig{Mode: common.TypeModeTest, RateLimit: 1})
			if err != nil {
				t.Fatalf("NewMetricsObserver() err = %v, want nil", err)
			}

			if tt.registerWith != nil {
				err := obs.Register(tt.registerWith)
				if tt.wantRegisterErr {
					if err == nil {
						t.Fatalf("Register() err = nil, want non-nil")
					}
				} else if err != nil {
					t.Fatalf("Register() err = %v, want nil", err)
				}
			}

			if tt.wantWorkerPoolAfter && obs.ReportWorkerPool == nil {
				t.Fatalf("ReportWorkerPool = nil, want non-nil")
			}

			if tt.notifyWith != nil {
				obs.Notify(tt.notifyWith)
			}
		})
	}
}

func TestMetricsObserver_Notify_WritesToFile(t *testing.T) {
	tmpDir := t.TempDir()
	auditPath := filepath.Join(tmpDir, "audit", "events.log")

	obs, err := NewMetricsObserver(MetricObserverConfig{
		Mode:      common.TypeModeTest,
		RateLimit: 1,
		FilePath:  auditPath,
	})
	if err != nil {
		t.Fatalf("NewMetricsObserver() err = %v, want nil", err)
	}
	t.Cleanup(func() {
		if obs.file != nil {
			_ = obs.file.Close()
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := obs.Register(ctx); err != nil {
		t.Fatalf("Register() err = %v, want nil", err)
	}

	ev := MetricsEvent{Ts: 123, IP: "10.0.0.1", Metrics: []string{"A", "B"}}
	obs.Notify(ev)

	deadline := time.Now().Add(750 * time.Millisecond)
	for {
		b, err := os.ReadFile(auditPath)
		if err == nil && len(b) > 0 {
			line := strings.TrimSpace(string(b))
			var got MetricsEvent
			if err := json.Unmarshal([]byte(line), &got); err != nil {
				t.Fatalf("unmarshal audit line err = %v; line=%q", err, line)
			}
			if got.Ts != ev.Ts || got.IP != ev.IP || strings.Join(got.Metrics, ",") != strings.Join(ev.Metrics, ",") {
				t.Fatalf("audit event = %#v, want %#v", got, ev)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting audit file to be written: %q", auditPath)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestMetricsObserver_Notify_SendsToServer(t *testing.T) {
	var called int32
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&called, 1)
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err == nil {
			gotBody = body
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	obs, err := NewMetricsObserver(MetricObserverConfig{
		Mode:      common.TypeModeTest,
		RateLimit: 1,
		URL:       srv.URL,
	})
	if err != nil {
		t.Fatalf("NewMetricsObserver() err = %v, want nil", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := obs.Register(ctx); err != nil {
		t.Fatalf("Register() err = %v, want nil", err)
	}

	ev := MetricsEvent{Ts: 555, IP: "1.2.3.4", Metrics: []string{"m"}}
	obs.Notify(ev)

	deadline := time.Now().Add(750 * time.Millisecond)
	for {
		if atomic.LoadInt32(&called) > 0 {
			var got MetricsEvent
			if err := json.Unmarshal(gotBody, &got); err != nil {
				t.Fatalf("unmarshal request body err = %v; body=%q", err, string(gotBody))
			}
			if got.Ts != ev.Ts || got.IP != ev.IP || strings.Join(got.Metrics, ",") != strings.Join(ev.Metrics, ",") {
				t.Fatalf("sent event = %#v, want %#v", got, ev)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting observer to call server")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
