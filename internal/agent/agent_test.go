package agent

import (
	"context"
	"sys-metrics/internal/config/agent"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestAgentConfig(serverURL string, poll, report time.Duration) *agent.Config {
	return agent.NewConfig(agent.InitProperties{
		PollInterval:   poll,
		ReportInterval: report,
		ServerAddr:     serverURL,
		Logger:         zap.NewExample(),
		RateLimit:      1,
	})
}

func TestAgent_Report(t *testing.T) {
	cfg := newTestAgentConfig(testServer.URL, 1, 1)
	a := NewAgent(cfg, context.Background())

	err := a.Report(context.Background())
	if err != nil {
		t.Fatalf("report failed %v", err)
	}
}

func TestAgent_StartPoll(t *testing.T) {
	cfg := newTestAgentConfig(testServer.URL, 1, 2)
	newAgent := NewAgent(cfg, context.Background())

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
	if len(newAgent.collector.Gauges) == 0 {
		t.Fatal("collector Gauges metrics is empty")
	}
	if len(newAgent.collector.Counters) == 0 {
		t.Fatal("collector Counters metrics is empty")
	}

}

func TestAgent_StartReport(t *testing.T) {
	cfg := newTestAgentConfig(testServer.URL, time.Second, time.Second)
	newAgent := NewAgent(cfg, context.Background())

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
	cfg := newTestAgentConfig("localhost", 2, 2)

	t.Run("creates agent with config", func(t *testing.T) {
		got := NewAgent(cfg, context.Background())

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
