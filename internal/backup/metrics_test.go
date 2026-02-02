package backup

import (
	"os"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"testing"
	"time"
)

func Test_writeMetricToBackup(t *testing.T) {
	type args struct {
		metric *metrics.Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "gauge metric",
			args: args{
				metric: &metrics.Metrics{
					ID:    "test_gauge",
					MType: "gauge",
					Value: func() *float64 { v := 123.45; return &v }(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewBackupConfig(common.TypeModeDevelopment, "./test", time.Second*10, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			err = config.Writer.WriteMetricToBackup(tt.args.metric)
			if err != nil {
				t.Fatalf("Failed to write metric to backup: %v", err)
			}
		})
	}
}

func TestBackupConfig_UpsertBatchMetricsToBackup(t *testing.T) {
	tests := []struct {
		name           string
		initialMetrics []metrics.Metrics
		newMetrics     []metrics.Metrics
		wantCount      int
		wantErr        bool
	}{
		{
			name:           "insert multiple new metrics",
			initialMetrics: []metrics.Metrics{},
			newMetrics: []metrics.Metrics{
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
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "update existing and add new",
			initialMetrics: []metrics.Metrics{
				{
					ID:    "gauge1",
					MType: "gauge",
					Value: func() *float64 { v := 10.0; return &v }(),
				},
			},
			newMetrics: []metrics.Metrics{
				{
					ID:    "gauge1",
					MType: "gauge",
					Value: func() *float64 { v := 20.0; return &v }(),
				},
				{
					ID:    "gauge2",
					MType: "gauge",
					Value: func() *float64 { v := 30.0; return &v }(),
				},
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "update all existing metrics",
			initialMetrics: []metrics.Metrics{
				{
					ID:    "gauge1",
					MType: "gauge",
					Value: func() *float64 { v := 10.0; return &v }(),
				},
				{
					ID:    "gauge2",
					MType: "gauge",
					Value: func() *float64 { v := 20.0; return &v }(),
				},
			},
			newMetrics: []metrics.Metrics{
				{
					ID:    "gauge1",
					MType: "gauge",
					Value: func() *float64 { v := 100.0; return &v }(),
				},
				{
					ID:    "gauge2",
					MType: "gauge",
					Value: func() *float64 { v := 200.0; return &v }(),
				},
			},
			wantCount: 2,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewBackupConfig(common.TypeModeTest, "./test", time.Second*10, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			defer config.Cleanup()

			if len(tt.initialMetrics) > 0 {
				if err := config.Writer.WriteBatchMetricsToBackup(tt.initialMetrics); err != nil {
					t.Fatalf("Failed to write initial metrics: %v", err)
				}
			}
			if err := config.UpsertBatchMetricsToBackup(tt.newMetrics); (err != nil) != tt.wantErr {
				t.Errorf("UpsertBatchMetricsToBackup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				config.Reader.Reset()
				resultMetrics, err := config.Reader.ReadFromBackup()
				if err != nil {
					t.Errorf("Failed to read metrics: %v", err)
					return
				}
				if len(resultMetrics) != tt.wantCount {
					t.Errorf("Expected %d metrics, got %d", tt.wantCount, len(resultMetrics))
				}
			}
		})
	}
}

func TestBackupConfig_UpsertMetricToBackup(t *testing.T) {
	tests := []struct {
		name           string
		initialMetrics []metrics.Metrics
		newMetric      *metrics.Metrics
		wantCount      int
		wantErr        bool
	}{
		{
			name:           "insert new metric",
			initialMetrics: []metrics.Metrics{},
			newMetric: &metrics.Metrics{
				ID:    "new_gauge",
				MType: "gauge",
				Value: func() *float64 { v := 100.0; return &v }(),
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "update existing metric",
			initialMetrics: []metrics.Metrics{
				{
					ID:    "test_gauge",
					MType: "gauge",
					Value: func() *float64 { v := 50.0; return &v }(),
				},
			},
			newMetric: &metrics.Metrics{
				ID:    "test_gauge",
				MType: "gauge",
				Value: func() *float64 { v := 150.0; return &v }(),
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "update one of multiple metrics",
			initialMetrics: []metrics.Metrics{
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
			},
			newMetric: &metrics.Metrics{
				ID:    "gauge1",
				MType: "gauge",
				Value: func() *float64 { v := 99.0; return &v }(),
			},
			wantCount: 2,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewBackupConfig(common.TypeModeTest, "./test", time.Second*10, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			defer config.Cleanup()
			if len(tt.initialMetrics) > 0 {
				if err := config.Writer.WriteBatchMetricsToBackup(tt.initialMetrics); err != nil {
					t.Fatalf("Failed to write initial metrics: %v", err)
				}
			}
			if err := config.UpsertMetricToBackup(tt.newMetric); (err != nil) != tt.wantErr {
				t.Errorf("UpsertMetricToBackup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				config.Reader.Reset()
				resultMetrics, err := config.Reader.ReadFromBackup()
				if err != nil {
					t.Errorf("Failed to read metrics: %v", err)
					return
				}
				if len(resultMetrics) != tt.wantCount {
					t.Errorf("Expected %d metrics, got %d", tt.wantCount, len(resultMetrics))
				}
				found := false
				for _, m := range resultMetrics {
					if m.ID == tt.newMetric.ID {
						found = true
						if m.MType != tt.newMetric.MType {
							t.Errorf("Metric type mismatch: got %s, want %s", m.MType, tt.newMetric.MType)
						}
						if tt.newMetric.Value != nil {
							if m.Value == nil || *m.Value != *tt.newMetric.Value {
								t.Errorf("Metric value mismatch")
							}
						}
						if tt.newMetric.Delta != nil {
							if m.Delta == nil || *m.Delta != *tt.newMetric.Delta {
								t.Errorf("Metric delta mismatch")
							}
						}
						break
					}
				}
				if !found {
					t.Errorf("Metric with ID %s not found in results", tt.newMetric.ID)
				}
			}
		})
	}
}

func TestWriter_WriteBatchMetricsToBackup(t *testing.T) {
	type args struct {
		metrics []metrics.Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "write batch of metrics",
			args: args{
				metrics: []metrics.Metrics{
					{
						ID:    "batch_gauge_1",
						MType: "gauge",
						Value: func() *float64 { v := 10.5; return &v }(),
					},
					{
						ID:    "batch_counter_1",
						MType: "counter",
						Delta: func() *int64 { v := int64(20); return &v }(),
					},
					{
						ID:    "batch_gauge_2",
						MType: "gauge",
						Value: func() *float64 { v := 30.75; return &v }(),
					},
					{
						ID:    "batch_counter_2",
						MType: "counter",
						Delta: func() *int64 { v := int64(40); return &v }(),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewBackupConfig(common.TypeModeTest, "./test", time.Second*10, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			defer config.Cleanup()
			w := config.Writer
			if err := w.WriteBatchMetricsToBackup(tt.args.metrics); (err != nil) != tt.wantErr {
				t.Errorf("WriteBatchMetricsToBackup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReader_ReadFromBackup(t *testing.T) {
	type fields struct {
		fileContent []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    []metrics.Metrics
		wantErr bool
	}{
		{
			name: "single gauge metric",
			fields: fields{
				fileContent: []byte(`{"id":"test_gauge","type":"gauge","value":123.45}`),
			},
			want: []metrics.Metrics{
				{
					ID:    "test_gauge",
					MType: "gauge",
					Value: func() *float64 { v := 123.45; return &v }(),
				},
			},
			wantErr: false,
		},
		{
			name: "single counter metric",
			fields: fields{
				fileContent: []byte(`{"id":"test_counter","type":"counter","delta":42}`),
			},
			want: []metrics.Metrics{
				{
					ID:    "test_counter",
					MType: "counter",
					Delta: func() *int64 { v := int64(42); return &v }(),
				},
			},
			wantErr: false,
		},
		{
			name: "multiple metrics",
			fields: fields{
				fileContent: []byte(`{"id":"test_gauge","type":"gauge","value":123.45}
{"id":"test_counter","type":"counter","delta":42}
{"id":"test_gauge2","type":"gauge","value":678.90}`),
			},
			want: []metrics.Metrics{
				{
					ID:    "test_gauge",
					MType: "gauge",
					Value: func() *float64 { v := 123.45; return &v }(),
				},
				{
					ID:    "test_counter",
					MType: "counter",
					Delta: func() *int64 { v := int64(42); return &v }(),
				},
				{
					ID:    "test_gauge2",
					MType: "gauge",
					Value: func() *float64 { v := 678.90; return &v }(),
				},
			},
			wantErr: false,
		},
		{
			name: "empty file",
			fields: fields{
				fileContent: []byte(``),
			},
			want:    []metrics.Metrics{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewBackupConfig(common.TypeModeTest, "./test", time.Second*10, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			defer func() {
				if err := config.Cleanup(); err != nil {
					t.Errorf("Failed to cleanup: %v", err)
				}
			}()

			err = os.WriteFile(config.getBackupFilename(), tt.fields.fileContent, 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			if err := config.Reader.Reset(); err != nil {
				t.Fatalf("Failed to reset reader: %v", err)
			}

			got, err := config.Reader.ReadFromBackup()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFromBackup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = config.Cleanup()
			if err != nil {
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ReadFromBackup() got %d metrics, want %d metrics", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].ID != tt.want[i].ID {
					t.Errorf("ReadFromBackup() metric[%d].ID = %v, want %v", i, got[i].ID, tt.want[i].ID)
				}
				if got[i].MType != tt.want[i].MType {
					t.Errorf("ReadFromBackup() metric[%d].MType = %v, want %v", i, got[i].MType, tt.want[i].MType)
				}
				if tt.want[i].Value != nil {
					if got[i].Value == nil {
						t.Errorf("ReadFromBackup() metric[%d].Value is nil, want %v", i, *tt.want[i].Value)
					} else if *got[i].Value != *tt.want[i].Value {
						t.Errorf("ReadFromBackup() metric[%d].Value = %v, want %v", i, *got[i].Value, *tt.want[i].Value)
					}
				}
				if tt.want[i].Delta != nil {
					if got[i].Delta == nil {
						t.Errorf("ReadFromBackup() metric[%d].Delta is nil, want %v", i, *tt.want[i].Delta)
					} else if *got[i].Delta != *tt.want[i].Delta {
						t.Errorf("ReadFromBackup() metric[%d].Delta = %v, want %v", i, *got[i].Delta, *tt.want[i].Delta)
					}
				}
			}
		})
	}
}
