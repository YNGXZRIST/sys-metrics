package proto

import (
	"testing"
)

func TestMetric_protoAPI(t *testing.T) {
	m := Metric_builder{
		Id:    "Alloc",
		Type:  Metric_GAUGE,
		Value: 3.14,
	}.Build()

	if m.GetId() != "Alloc" {
		t.Fatalf("id = %q", m.GetId())
	}
	if m.GetType() != Metric_GAUGE {
		t.Fatalf("type = %v", m.GetType())
	}
	if m.GetValue() != 3.14 {
		t.Fatalf("value = %v", m.GetValue())
	}

	m.SetId("Heap")
	m.SetType(Metric_COUNTER)
	m.SetDelta(7)
	m.SetValue(0)

	if m.GetDelta() != 7 {
		t.Fatalf("delta = %d", m.GetDelta())
	}

	_ = m.String()
	m.Reset()
	_ = m.ProtoReflect().Descriptor().FullName()

	enum := Metric_GAUGE.Enum()
	if enum == nil || *enum != Metric_GAUGE {
		t.Fatal("enum")
	}
	if Metric_GAUGE.String() == "" {
		t.Fatal("empty enum string")
	}
	_ = Metric_GAUGE.Descriptor()
	_ = Metric_GAUGE.Type()
	_ = Metric_GAUGE.Number()
}

func TestMetric_counterProto(t *testing.T) {
	m := Metric_builder{Id: "PollCount", Type: Metric_COUNTER, Delta: 42}.Build()
	if m.GetType() != Metric_COUNTER || m.GetDelta() != 42 {
		t.Fatalf("counter: type=%v delta=%d", m.GetType(), m.GetDelta())
	}
	_ = Metric_COUNTER.String()
}

func TestUpdateMetricsRequest_protoAPI(t *testing.T) {
	req := UpdateMetricsRequest_builder{
		Metrics: []*Metric{
			Metric_builder{Id: "Sys", Type: Metric_GAUGE, Value: 1}.Build(),
		},
	}.Build()

	if len(req.GetMetrics()) != 1 {
		t.Fatalf("metrics len = %d", len(req.GetMetrics()))
	}

	resp := UpdateMetricsResponse_builder{}.Build()
	_ = resp.String()
	resp.Reset()
	_ = resp.ProtoReflect()
}
