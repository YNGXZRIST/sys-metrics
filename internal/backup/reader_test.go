package backup

import (
	"os"
	"sys-metrics/internal/common"
	"sys-metrics/pkg/filesystem"
	"testing"
	"time"
)

func TestNewBackupReader(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    *Reader
		wantErr bool
	}{
		{
			name: "create reader for existing file",
			args: args{
				filename: "./test/backup_reader_test.metrics",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "create reader for non-existent file",
			args: args{
				filename: "/tmp/absolutely_non_existent_file_12345.metrics",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				dir := "./test"
				err := filesystem.CreateDirIfNotExists(dir)
				if err != nil {
					t.Fatalf("Failed to create dir: %v", err)
				}
				err = os.WriteFile(tt.args.filename, []byte(`{"id":"test","type":"gauge","value":1.0}`), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer os.Remove(tt.args.filename)
			}

			got, err := newBackupReader(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("newBackupReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Errorf("newBackupReader() returned nil reader")
					return
				}
				if got.file == nil {
					t.Errorf("newBackupReader() file is nil")
				}
				if got.reader == nil {
					t.Errorf("newBackupReader() reader is nil")
				}
				_ = got.Close()
			}
		})
	}
}

func TestReader_Close(t *testing.T) {
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
			config, err := NewBackupConfig(common.TypeModeTest, "./test", time.Second*10, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			defer config.Cleanup()
			if err := config.Reader.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
