package agent

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/agent"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestAgent_Report(t *testing.T) {
	cfg := agent.NewConfig(1, 1, testServer.URL, zap.NewExample())
	a := NewAgent(cfg)

	err := a.Report()
	if err != nil {
		t.Fatalf("report failed %v", err)
	}
}

func TestAgent_StartPoll(t *testing.T) {
	cfg := agent.NewConfig(1, 2, testServer.URL, zap.NewExample())
	newAgent := NewAgent(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		newAgent.StartPoll(ctx)
		close(done)
	}()
	time.Sleep(1500 * time.Millisecond)

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("StartPoll not ended by context")
	}
	if newAgent.collector == nil {
		t.Fatal("collector is nil")
	}
	if len(newAgent.collector.metrics) == 0 {
		t.Fatal("collector metrics is empty")
	}
	if len(newAgent.collector.metrics[common.Gauge]) == 0 {
		t.Fatal("metrics Gauge is empty")
	}
	if len(newAgent.collector.metrics[common.Counter]) == 0 {
		t.Fatal("metrics Counter is empty")
	}
}

func TestAgent_StartReport(t *testing.T) {
	cfg := agent.NewConfig(1*time.Second, 1*time.Second, testServer.URL, zap.NewExample())
	newAgent := NewAgent(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		newAgent.StartReport(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("StartReport not ended by context")
	}
}

func TestNewAgent(t *testing.T) {
	cfg := agent.NewConfig(2, 2, "localhost", zap.NewExample())

	t.Run("creates agent with config", func(t *testing.T) {
		got := NewAgent(cfg)

		if got == nil {
			t.Fatal("NewAgent() returned nil")
		}
		if got.Config != cfg {
			t.Errorf("NewAgent().Config = %v, want %v", got.Config, cfg)
		}
		if got.collector == nil {
			t.Error("NewAgent().collector should not be nil")
		}
	})
}
