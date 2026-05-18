package agent

import (
	"context"
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	"testing"
)

func TestCollector_GetGauge_GetCounter(t *testing.T) {
	c := NewCollector(context.Background(), 1)
	g := models.NewGauge("g1")
	g.SetValue(3.5)
	c.Gauges["g1"] = g

	got, ok := c.GetGauge("g1")
	if !ok || got == nil || *got.Value != 3.5 {
		t.Fatalf("GetGauge: ok=%v val=%v", ok, got)
	}
	got2, ok2 := c.GetGauge("missing")
	if ok2 || got2 == nil {
		t.Fatal("expected synthetic gauge for missing name")
	}

	c.SetPollCounterMetric()
	pc := c.GetPollCountMetric()
	if pc == nil || pc.Delta == nil {
		t.Fatal("poll counter")
	}

	co, ok := c.GetCounter(common.PollCount)
	if !ok || co == nil {
		t.Fatalf("GetCounter poll: ok=%v", ok)
	}
}
