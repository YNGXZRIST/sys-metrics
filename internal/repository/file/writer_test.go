package file

import (
	"context"
	"os"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"testing"
	"time"
)

func Test_newBackupWriter(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    *Writer
		wantErr bool
	}{
		{
			name: "create writer for new file",
			args: args{
				filename: "./test/backup_writer_test_new.metrics",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "create writer for existing file (append mode)",
			args: args{
				filename: "./test/backup_writer_test_exist.metrics",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "create writer for file in non-existent directory",
			args: args{
				filename: "/tmp/non_existent_dir_12345/backup.metrics",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				dir := "./test"
				err := os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					t.Fatalf("Failed to create dir: %v", err)
				}
				if tt.name == "create writer for existing file (append mode)" {
					err = os.WriteFile(tt.args.filename, []byte(`existing content`), 0644)
					if err != nil {
						t.Fatalf("Failed to create existing file: %v", err)
					}
				}

				defer os.Remove(tt.args.filename)
			}

			got, err := newBackupWriter(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("newBackupWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Errorf("newBackupWriter() returned nil writer")
					return
				}
				if got.file == nil {
					t.Errorf("newBackupWriter() file is nil")
				}
				if got.writer == nil {
					t.Errorf("newBackupWriter() writer is nil")
				}
				testData := []byte("test data\n")
				_, writeErr := got.writer.Write(testData)
				if writeErr != nil {
					t.Errorf("Failed to write test data: %v", writeErr)
				}

				_ = got.Close()
			}
		})
	}
}

func TestWriter_Close(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful close",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(common.TypeModeTest, "./test", time.Second*10, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			defer config.Cleanup()
			storage, err := NewMetricFileBackupStorage(config)
			if err != nil {
				t.Fatalf("Failed to create backup storage: %v", err)
			}
			defer func() { _ = storage.Reader.Close() }()

			testMetric := &metrics.Metrics{
				ID:    "test_close",
				MType: "gauge",
				Value: func() *float64 { v := 42.0; return &v }(),
			}
			err = storage.MetricsHandler.Write(context.Background(), testMetric)
			if err != nil {
				t.Fatalf("Failed to write metric: %v", err)
			}

			if err := storage.Writer.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
