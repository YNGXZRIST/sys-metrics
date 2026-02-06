package repository

import (
	"os"
	"sys-metrics/internal/model/metrics"
	"testing"
)

func createTempBackupFile(t *testing.T) (string, func()) {
	file, err := os.CreateTemp("", "backup_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return file.Name(), func() { _ = os.Remove(file.Name()) }
}

func TestFileMetricsBackupHandler_Write_Read(t *testing.T) {
	filename, cleanup := createTempBackupFile(t)
	defer cleanup()
	reader, err := NewBackupReader(filename)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	writer, err := NewBackupWriter(filename)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	h := NewFileMetricsBackupHandler(reader, writer)
	metric := metrics.Metrics{
		ID:    "test_gauge",
		MType: "gauge",
		Value: func() *float64 { v := 123.45; return &v }(),
	}
	if err := h.Write(&metric); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := reader.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	got, err := h.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != metric.ID || got[0].MType != metric.MType || *got[0].Value != *metric.Value {
		t.Errorf("Read() got = %v, want %v", got, metric)
	}
}

func TestFileMetricsBackupHandler_WriteBatch_Read(t *testing.T) {
	filename, cleanup := createTempBackupFile(t)
	defer cleanup()
	reader, err := NewBackupReader(filename)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	writer, err := NewBackupWriter(filename)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	h := NewFileMetricsBackupHandler(reader, writer)
	metricsBatch := []metrics.Metrics{
		{
			ID:    "gauge1",
			MType: "gauge",
			Value: func() *float64 { v := 10.0; return &v }(),
		},
		{
			ID:    "counter1",
			MType: "counter",
			Delta: func() *int64 { v := int64(5); return &v }(),
		},
	}
	if err := h.WriteBatch(metricsBatch); err != nil {
		t.Fatalf("WriteBatch() error = %v", err)
	}
	if err := reader.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	got, err := h.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("Read() got %d metrics, want 2", len(got))
	}
}

func TestFileMetricsBackupHandler_Upsert(t *testing.T) {
	filename, cleanup := createTempBackupFile(t)
	defer cleanup()
	reader, err := NewBackupReader(filename)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	writer, err := NewBackupWriter(filename)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	h := NewFileMetricsBackupHandler(reader, writer)
	initial := metrics.Metrics{
		ID:    "gauge1",
		MType: "gauge",
		Value: func() *float64 { v := 10.0; return &v }(),
	}
	if err := h.Write(&initial); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	updated := metrics.Metrics{
		ID:    "gauge1",
		MType: "gauge",
		Value: func() *float64 { v := 20.0; return &v }(),
	}
	if err := h.Upsert(&updated); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if err := reader.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	got, err := h.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(got) != 1 || *got[0].Value != 20.0 {
		t.Errorf("Upsert() did not update metric value, got %v", got)
	}
}

func TestFileMetricsBackupHandler_UpsertBatch(t *testing.T) {
	filename, cleanup := createTempBackupFile(t)
	defer cleanup()
	reader, err := NewBackupReader(filename)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	writer, err := NewBackupWriter(filename)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	h := NewFileMetricsBackupHandler(reader, writer)
	initial := []metrics.Metrics{
		{
			ID:    "gauge1",
			MType: "gauge",
			Value: func() *float64 { v := 10.0; return &v }(),
		},
		{
			ID:    "counter1",
			MType: "counter",
			Delta: func() *int64 { v := int64(5); return &v }(),
		},
	}
	if err := h.WriteBatch(initial); err != nil {
		t.Fatalf("WriteBatch() error = %v", err)
	}
	batch := []metrics.Metrics{
		{
			ID:    "gauge1",
			MType: "gauge",
			Value: func() *float64 { v := 100.0; return &v }(),
		},
		{
			ID:    "counter2",
			MType: "counter",
			Delta: func() *int64 { v := int64(10); return &v }(),
		},
	}
	if err := h.UpsertBatch(batch); err != nil {
		t.Fatalf("UpsertBatch() error = %v", err)
	}
	if err := reader.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	got, err := h.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("UpsertBatch() got %d metrics, want 3", len(got))
	}
	found := false
	for _, m := range got {
		if m.ID == "gauge1" && m.Value != nil && *m.Value == 100.0 {
			found = true
		}
	}
	if !found {
		t.Errorf("UpsertBatch() did not update gauge1 value")
	}
}
