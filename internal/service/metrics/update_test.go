package metrics

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/middleware"
	models "sys-metrics/internal/model/metrics"
	"sys-metrics/internal/observer"
	"sys-metrics/internal/repository/memory"
	"sys-metrics/internal/repository/metrics"
	"testing"
)

type captureNotifier struct {
	events []any
}

func (n *captureNotifier) Notify(ctx context.Context, data any) {
	_ = ctx
	n.events = append(n.events, data)
}

func TestUpdate(t *testing.T) {
	type args struct {
		metricType string
		name       string
		value      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "empty",
			args:    args{},
			wantErr: true,
		},
		{
			name: "unknown type",
			args: args{
				metricType: "histogram",
				name:       "x",
				value:      "1",
			},
			wantErr: true,
		},
		{
			name: "bad counter value",
			args: args{
				metricType: common.Counter,
				name:       "c",
				value:      "nope",
			},
			wantErr: true,
		},
		{
			name: "bad gauge value",
			args: args{
				metricType: "gauge",
				name:       "g",
				value:      "x.y.z",
			},
			wantErr: true,
		},
		{
			name: "metric counter",
			args: args{
				metricType: common.Counter,
				name:       "c1",
				value:      "42",
			},
			wantErr: false,
		},
		{
			name: "metric gauge",
			args: args{
				metricType: "gauge",
				name:       "test",
				value:      "10",
			},
			wantErr: false,
		},
	}
	metrics.Init(memory.NewService())
	ms := NewService(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ms.Update(context.Background(), tt.args.metricType, tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricService_UpdateMetric(t *testing.T) {
	metrics.Init(memory.NewService())
	notifier := &captureNotifier{}
	ms := NewService(notifier)
	ctx := context.WithValue(context.Background(), middleware.CtxClientIPKey, "127.0.0.1")

	v := 12.5
	got, err := ms.UpdateMetric(ctx, models.Metrics{
		ID:    "json_gauge",
		MType: common.Gauge,
		Value: &v,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Value == nil || *got.Value != v {
		t.Fatalf("UpdateMetric() value = %v, want %v", got.Value, v)
	}
	if len(notifier.events) != 1 {
		t.Fatalf("Notify calls = %d, want 1", len(notifier.events))
	}
	event, ok := notifier.events[0].(observer.MetricsEvent)
	if !ok {
		t.Fatalf("event type = %T, want observer.MetricsEvent", notifier.events[0])
	}
	if event.IP != "127.0.0.1" || len(event.Metrics) != 1 || event.Metrics[0] != "json_gauge" {
		t.Fatalf("event = %+v", event)
	}
}

func TestMetricService_UpdateMetricCounterDefault(t *testing.T) {
	metrics.Init(memory.NewService())
	ms := NewService(nil)

	got, err := ms.UpdateMetric(context.Background(), models.Metrics{
		ID:    "json_counter",
		MType: common.Counter,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Delta == nil || *got.Delta != 0 {
		t.Fatalf("UpdateMetric() delta = %v, want 0", got.Delta)
	}
}

func TestMetricService_UpdateMetricUnknownType(t *testing.T) {
	metrics.Init(memory.NewService())
	ms := NewService(nil)

	_, err := ms.UpdateMetric(context.Background(), models.Metrics{
		ID:    "x",
		MType: "histogram",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMetricService_GetMetric(t *testing.T) {
	metrics.Init(memory.NewService())
	ms := NewService(nil)
	ctx := context.Background()
	counter := models.NewCounter("stored_counter")
	counter.SetValue(7)
	if err := metrics.Counters().Set(ctx, counter.ID, counter); err != nil {
		t.Fatal(err)
	}
	gauge := models.NewGauge("stored_gauge")
	gauge.SetValue(9.5)
	if err := metrics.Gauges().Set(ctx, gauge.ID, gauge); err != nil {
		t.Fatal(err)
	}

	gotCounter, err := ms.GetMetric(ctx, common.Counter, counter.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotCounter.Delta == nil || *gotCounter.Delta != 7 {
		t.Fatalf("counter = %+v", gotCounter)
	}

	gotGauge, err := ms.GetMetric(ctx, common.Gauge, gauge.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotGauge.Value == nil || *gotGauge.Value != 9.5 {
		t.Fatalf("gauge = %+v", gotGauge)
	}

	if _, err := ms.GetMetric(ctx, "unknown", "x"); err == nil {
		t.Fatal("expected unknown type error")
	}
	if _, err := ms.GetMetric(ctx, common.Gauge, "missing"); err == nil {
		t.Fatal("expected missing metric error")
	}
}

func TestMetricService_MetricValue(t *testing.T) {
	ms := NewService(nil)
	gaugeValue := 1.25
	counterValue := int64(4)

	gotGauge, err := ms.MetricValue(models.Metrics{MType: common.Gauge, Value: &gaugeValue})
	if err != nil || gotGauge != "1.25" {
		t.Fatalf("MetricValue(gauge) = %q, %v", gotGauge, err)
	}

	gotCounter, err := ms.MetricValue(models.Metrics{MType: common.Counter, Delta: &counterValue})
	if err != nil || gotCounter != "4" {
		t.Fatalf("MetricValue(counter) = %q, %v", gotCounter, err)
	}

	if _, err := ms.MetricValue(models.Metrics{MType: "bad"}); err == nil {
		t.Fatal("expected unknown type error")
	}
}

func Test_updateCounter(t *testing.T) {
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				name:  "test",
				value: 12,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics.Init(memory.NewService())
			if err := updateCounter(context.Background(), tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("updateCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
			val, err := metrics.Counters().Get(context.Background(), tt.args.name)
			if err != nil {
				t.Errorf("counters.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if *val.Delta != tt.args.value {
				t.Errorf("updateCounter() delta got %v, want %v", *val.Delta, tt.args.value)
			}
		})
	}
}

func Test_updateGauge(t *testing.T) {
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				name:  "test",
				value: 12,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics.Init(memory.NewService())
			if err := updateGauge(context.Background(), tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("updateGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
			val, err := metrics.Gauges().Get(context.Background(), tt.args.name)
			if err != nil {
				t.Errorf("updateGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
			if *val.Value != tt.args.value {
				t.Errorf("updateGauge() Value got %v, want %v", *val.Delta, tt.args.value)
			}
		})
	}
}

func TestBatchUpdateMetrics_empty(t *testing.T) {
	metrics.Init(memory.NewService())
	ms := NewService(nil)
	if err := ms.BatchUpdateMetrics(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := ms.BatchUpdateMetrics(context.Background(), []models.Metrics{}); err != nil {
		t.Fatal(err)
	}
}

func TestBatchUpdateMetrics_oneGauge(t *testing.T) {
	metrics.Init(memory.NewService())
	ms := NewService(nil)
	v := 3.0
	err := ms.BatchUpdateMetrics(context.Background(), []models.Metrics{
		{ID: "g1", MType: "gauge", Value: &v},
	})
	if err != nil {
		t.Fatal(err)
	}
	g, err := metrics.Gauges().Get(context.Background(), "g1")
	if err != nil || g.Value == nil || *g.Value != v {
		t.Fatalf("gauge: err=%v val=%v", err, g.Value)
	}
}
