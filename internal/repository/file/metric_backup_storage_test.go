package file

import (
	"testing"
)

func TestMetricBackupStorage_NeedSync(t *testing.T) {
	type fields struct {
		BackupStorage *BackupStorage
		Config        *Config
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Sync mode (interval=0)",
			fields: fields{
				BackupStorage: &BackupStorage{},
				Config:        &Config{Interval: 0},
			},
			want: true,
		},
		{
			name: "Async mode (interval>0)",
			fields: fields{
				BackupStorage: &BackupStorage{},
				Config:        &Config{Interval: 1},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MetricBackupStorage{
				BackupStorage: tt.fields.BackupStorage,
				Config:        tt.fields.Config,
			}
			if got := s.NeedSync(); got != tt.want {
				t.Errorf("NeedSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMetricFileBackupStorage(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid config",
			args: args{
				config: &Config{},
			},
			wantErr: false,
		},
		{
			name: "Nil config",
			args: args{
				config: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMetricFileBackupStorage(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMetricFileBackupStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("NewMetricFileBackupStorage() got = nil, want not nil")
			}
		})
	}
}
