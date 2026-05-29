package file

import (
	"context"
	"os"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/metricsiface"
	"testing"
	"time"
)

func testBackupHandler(t *testing.T) (metricsiface.Handler, *BackupHandler, func()) {
	t.Helper()
	cfg, err := NewConfig(common.TypeModeTest, "", time.Second, true)
	if err != nil {
		t.Fatal(err)
	}
	storage, err := NewMetricFileBackupStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	bh := storage.MetricsHandler.(*BackupHandler)
	return storage.MetricsHandler, bh, func() {
		_ = storage.Close()
		_ = cfg.Close()
		_ = cfg.Cleanup()
	}
}

func TestBackupHandler_Upsert(t *testing.T) {
	h, bh, cleanup := testBackupHandler(t)
	defer cleanup()

	ctx := context.Background()
	g := metrics.NewGauge("g1")
	g.SetValue(1.5)
	if err := h.Upsert(ctx, &g.Metrics); err != nil {
		t.Fatal(err)
	}

	g2 := metrics.NewGauge("g1")
	g2.SetValue(2.5)
	if err := h.Upsert(ctx, &g2.Metrics); err != nil {
		t.Fatal(err)
	}

	if err := bh.reader.Reset(); err != nil {
		t.Fatal(err)
	}
	got, err := h.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Value == nil || *got[0].Value != 2.5 {
		t.Fatalf("metrics = %+v", got)
	}
}

func TestBackupHandler_UpsertBatch(t *testing.T) {
	_, bh, cleanup := testBackupHandler(t)
	defer cleanup()

	ctx := context.Background()
	c := metrics.NewCounter("c1")
	c.SetValue(10)
	g := metrics.NewGauge("g1")
	g.SetValue(4.2)

	if err := bh.UpsertBatch(ctx, []metrics.Metrics{c.Metrics, g.Metrics}); err != nil {
		t.Fatal(err)
	}

	c2 := metrics.NewCounter("c1")
	c2.SetValue(20)
	if err := bh.UpsertBatch(ctx, []metrics.Metrics{c2.Metrics}); err != nil {
		t.Fatal(err)
	}

	if err := bh.reader.Reset(); err != nil {
		t.Fatal(err)
	}
	got, err := bh.Read(ctx)
	if err != nil || len(got) != 2 {
		t.Fatalf("read: len=%d err=%v", len(got), err)
	}
}

func TestBackupService_WriteBatchMetrics_andClose(t *testing.T) {
	cfg, err := NewConfig(common.TypeModeTest, "", time.Second, true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = cfg.Close()
		_ = cfg.Cleanup()
	}()
	storage, err := NewMetricFileBackupStorage(cfg)
	if err != nil {
		t.Fatal(err)
	}
	svc := NewBackupService(storage)

	ctx := context.Background()
	v := 9.9
	d := int64(3)
	if err := svc.WriteBatchMetrics(ctx, []metrics.Metrics{
		{ID: "g", MType: common.Gauge, Value: &v},
		{ID: "c", MType: common.Counter, Delta: &d},
	}); err != nil {
		t.Fatal(err)
	}
	if err := svc.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestConfig_Cleanup_removesTempDir(t *testing.T) {
	cfg, err := NewConfig(common.TypeModeTest, "", time.Second, true)
	if err != nil {
		t.Fatal(err)
	}
	dir := cfg.StoragePath
	if err := cfg.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("dir still exists: %v", err)
	}
}
