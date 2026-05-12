package metrics

import (
	"encoding/json"
	"sys-metrics/internal/common"
	"testing"
)

func TestMetrics_JSON_roundTrip(t *testing.T) {
	v := 1.25
	m := Metrics{
		ID:    "x",
		MType: common.Gauge,
		Value: &v,
		Hash:  "h",
	}
	data, err := json.Marshal(&m)
	if err != nil {
		t.Fatal(err)
	}
	var out Metrics
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != m.ID || out.MType != m.MType || out.Hash != m.Hash {
		t.Fatalf("meta mismatch: %+v", out)
	}
	if out.Value == nil || *out.Value != v {
		t.Fatalf("value = %v", out.Value)
	}
}

func TestMetrics_JSON_counter(t *testing.T) {
	d := int64(7)
	m := Metrics{ID: "c", MType: common.Counter, Delta: &d}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var out Metrics
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.Delta == nil || *out.Delta != d {
		t.Fatal("delta mismatch")
	}
}
