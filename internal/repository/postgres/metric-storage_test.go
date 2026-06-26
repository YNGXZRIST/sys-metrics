//go:build integration
// +build integration

package postgres

import (
	"context"
	"reflect"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/model/metrics"
	"testing"
)

func TestNewMetricStorage(t *testing.T) {
	t.Run("nil db", func(t *testing.T) {
		got := NewMetricStorage(nil)
		if got == nil {
			t.Fatal("NewMetricStorage(nil) returned nil")
		}
		if got.Config == nil {
			t.Error("NewMetricStorage(nil).Config is nil")
		}
		if got.Config != nil && got.Config.conn != nil {
			t.Error("NewMetricStorage(nil).Config.conn should be nil")
		}
		if got.BackupStorage == nil {
			t.Error("NewMetricStorage(nil).BackupStorage is nil")
		}
	})

	t.Run("with conn", func(t *testing.T) {
		conn := getTestDB(t)
		got := NewMetricStorage(conn)
		if got == nil {
			t.Fatal("NewMetricStorage(conn) returned nil")
		}
		if got.Config == nil || got.Config.conn != conn {
			t.Error("NewMetricStorage(conn).Config.conn != conn")
		}
		if got.BackupStorage == nil {
			t.Error("NewMetricStorage(conn).BackupStorage is nil")
		}
	})
}

func TestMetricStorage_Close(t *testing.T) {
	if testDSN == "" {
		t.Skip("Test DB not available")
	}
	conn, err := db.NewConn(db.NewCfg(&db.Config{DNS: testDSN}))
	if err != nil {
		t.Fatalf("create conn for Close test: %v", err)
	}
	storage := NewMetricStorage(conn)
	ctx := context.Background()
	err = storage.Close(ctx)
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestMetricStorage_Counters(t *testing.T) {
	conn := getTestDB(t)
	storage := NewMetricStorage(conn)
	ctx := context.Background()
	got := storage.Counters()
	if got == nil {
		t.Fatal("Counters() returned nil")
	}
	counter := metrics.NewCounter(common.PollCount)
	counter.SetValue(1)
	err := got.Set(ctx, counter.ID, counter)
	if err != nil {
		t.Errorf("Counters().Set() error = %v", err)
	}
	all := got.All(ctx)
	if len(all) == 0 {
		t.Error("Counters().All() empty after Set")
	}
	if g, ok := all[counter.ID]; !ok || g.Metrics.Delta == nil || *g.Metrics.Delta != 1 {
		t.Errorf("Counters().All()[%q] = %+v", counter.ID, all[counter.ID])
	}
}

func TestMetricStorage_Gauges(t *testing.T) {
	conn := getTestDB(t)
	storage := NewMetricStorage(conn)
	ctx := context.Background()
	got := storage.Gauges()
	if got == nil {
		t.Fatal("Gauges() returned nil")
	}
	gauge := metrics.NewGauge(common.Alloc)
	gauge.SetValue(1.5)
	err := got.Set(ctx, gauge.ID, gauge)
	if err != nil {
		t.Errorf("Gauges().Set() error = %v", err)
	}
	all := got.All(ctx)
	if len(all) == 0 {
		t.Error("Gauges().All() empty after Set")
	}
	if g, ok := all[gauge.ID]; !ok || g.Metrics.Value == nil || *g.Metrics.Value != 1.5 {
		t.Errorf("Gauges().All()[%q] = %+v", gauge.ID, all[gauge.ID])
	}
}

func TestMetricStorage_GetAllMetrics(t *testing.T) {
	conn := getTestDB(t)
	storage := NewMetricStorage(conn)
	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		got := storage.GetAllMetrics(ctx)
		if got == nil {
			t.Fatal("GetAllMetrics() returned nil")
		}
		if len(got) != 0 {
			t.Errorf("GetAllMetrics() empty storage: got %d, want 0", len(got))
		}
	})

	t.Run("with metrics", func(t *testing.T) {
		gauge := metrics.NewGauge(common.Alloc)
		gauge.SetValue(1)
		counter := metrics.NewCounter(common.PollCount)
		counter.SetValue(2)
		_ = storage.Gauges().Set(ctx, gauge.ID, gauge)
		_ = storage.Counters().Set(ctx, counter.ID, counter)
		got := storage.GetAllMetrics(ctx)
		if len(got) != 2 {
			t.Fatalf("GetAllMetrics() got %d, want 2", len(got))
		}
		byID := make(map[string]metrics.Metrics)
		for _, m := range got {
			byID[m.ID] = m
		}
		if !reflect.DeepEqual(byID[gauge.ID], gauge.Metrics) {
			t.Errorf("gauge: got %+v, want %+v", byID[gauge.ID], gauge.Metrics)
		}
		if !reflect.DeepEqual(byID[counter.ID], counter.Metrics) {
			t.Errorf("counter: got %+v, want %+v", byID[counter.ID], counter.Metrics)
		}
	})
}

func TestMetricStorage_InitRoutine(t *testing.T) {
	conn := getTestDB(t)
	storage := NewMetricStorage(conn)
	ctx := context.Background()
	err := storage.InitRoutine(ctx)
	if err != nil {
		t.Errorf("InitRoutine() error = %v", err)
	}
}

func TestMetricStorage_ReadBackup(t *testing.T) {
	conn := getTestDB(t)
	ctx := context.Background()
	gauge := metrics.NewGauge(common.Alloc)
	gauge.SetValue(2.5)
	counter := metrics.NewCounter(common.PollCount)
	counter.SetValue(3)
	cfg := NewConfig(conn)
	err := cfg.handler.WriteBatch(ctx, []metrics.Metrics{gauge.Metrics, counter.Metrics})
	if err != nil {
		t.Errorf("WriteBatch() error = %v", err)
	}
	storage := NewMetricStorage(conn)
	err = storage.ReadBackup(ctx)
	if err != nil {
		t.Fatalf("ReadBackup() error = %v", err)
	}
	got := storage.GetAllMetrics(ctx)
	if len(got) < 2 {
		t.Fatalf("after ReadBackup got %d metrics, want at least 2", len(got))
	}
	byID := make(map[string]metrics.Metrics)
	for _, m := range got {
		byID[m.ID] = m
	}
	if g, ok := byID[gauge.ID]; !ok || g.Value == nil || *g.Value != 2.5 {
		t.Errorf("after ReadBackup gauge: got %+v, want value 2.5", byID[gauge.ID])
	}
	if c, ok := byID[counter.ID]; !ok || c.Delta == nil || *c.Delta != 3 {
		t.Errorf("after ReadBackup counter: got %+v, want delta 3", byID[counter.ID])
	}
}

func TestMetricStorage_WriteBackup(t *testing.T) {
	conn := getTestDB(t)
	storage := NewMetricStorage(conn)
	ctx := context.Background()

	gauge := metrics.NewGauge(common.Alloc)
	gauge.SetValue(3)
	counter := metrics.NewCounter(common.PollCount)
	counter.SetValue(4)
	err := storage.Gauges().Set(ctx, gauge.ID, gauge)
	if err != nil {
		t.Errorf("set gauge  error = %v", err)
	}
	err = storage.Counters().Set(ctx, counter.ID, counter)
	if err != nil {
		t.Errorf("set counter  error = %v", err)
	}
	err = storage.WriteBackup(ctx)
	if err != nil {
		t.Fatalf("WriteBackup() error = %v", err)
	}
	read, err := storage.Config.handler.Read(ctx)
	if err != nil {
		t.Fatalf("Read after WriteBackup: %v", err)
	}
	byID := make(map[string]metrics.Metrics)
	for _, m := range read {
		byID[m.ID] = m
	}
	if g, ok := byID[gauge.ID]; !ok || g.Value == nil || *g.Value != 3 {
		t.Errorf("after WriteBackup gauge in DB: got %+v", byID[gauge.ID])
	}
	if c, ok := byID[counter.ID]; !ok || c.Delta == nil || *c.Delta != 4 {
		t.Errorf("after WriteBackup counter in DB: got %+v", byID[counter.ID])
	}
}
