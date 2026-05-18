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
		registerWith        any
		notifyWith          any
		name                string
		wantRegisterErr     bool
		wantWorkerPoolAfter bool
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
			notifyWith: MetricsEvent{TS: 1, IP: "127.0.0.1", Metrics: []string{"m1"}},
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
			notifyWith:          MetricsEvent{TS: 2, IP: "127.0.0.2", Metrics: []string{"m2"}},
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
				obs.Notify(ctx, tt.notifyWith)
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

	ev := MetricsEvent{TS: 123, IP: "10.0.0.1", Metrics: []string{"A", "B"}}
	obs.Notify(ctx, ev)

	deadline := time.Now().Add(750 * time.Millisecond)
	for {
		b, err := os.ReadFile(auditPath)
		if err == nil && len(b) > 0 {
			line := strings.TrimSpace(string(b))
			var got MetricsEvent
			if err := json.Unmarshal([]byte(line), &got); err != nil {
				t.Fatalf("unmarshal audit line err = %v; line=%q", err, line)
			}
			if got.TS != ev.TS || got.IP != ev.IP || strings.Join(got.Metrics, ",") != strings.Join(ev.Metrics, ",") {
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
	received := make(chan []byte, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// copy: handler runs on server goroutine; channel handoff avoids races with the test.
		received <- append([]byte(nil), body...)
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

	ev := MetricsEvent{TS: 555, IP: "1.2.3.4", Metrics: []string{"m"}}
	obs.Notify(ctx, ev)

	select {
	case gotBody := <-received:
		var got MetricsEvent
		if err := json.Unmarshal(gotBody, &got); err != nil {
			t.Fatalf("unmarshal request body err = %v; body=%q", err, string(gotBody))
		}
		if got.TS != ev.TS || got.IP != ev.IP || strings.Join(got.Metrics, ",") != strings.Join(ev.Metrics, ",") {
			t.Fatalf("sent event = %#v, want %#v", got, ev)
		}
	case <-time.After(750 * time.Millisecond):
		t.Fatal("timeout waiting observer to call server")
	}
}

func TestSaveToFilePath_writesLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	obs, err := NewMetricsObserver(MetricObserverConfig{Mode: common.TypeModeTest, RateLimit: 1, FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	if obs.file != nil {
		defer obs.file.Close()
	}
	ev := MetricsEvent{TS: 9, IP: "10.0.0.1", Metrics: []string{"a", "b"}}
	obs.SaveToFilePath(ev)
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), `"metrics":["a","b"]`) {
		t.Fatalf("file: %q err=%v", data, err)
	}
}
