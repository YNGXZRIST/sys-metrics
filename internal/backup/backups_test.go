package backup

import (
	"os"
	"sys-metrics/internal/common"
	"testing"
	"time"
)

func TestBackupConfig_initBackupRoutine(t *testing.T) {
	type fields struct {
		fileContent []byte
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "init backup routine",
			fields: fields{
				fileContent: []byte(`{"ID":"test_gauge","MType":"gauge","Value":123.45}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewBackupConfig(common.TypeModeTest, "./test", time.Second*10, true)
			if err != nil {
				t.Fatalf("Failed to create backup config: %v", err)
			}
			err = os.WriteFile(config.getBackupFilename(), tt.fields.fileContent, 0644)

		})
	}
}
