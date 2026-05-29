package config

import (
	"reflect"
	"testing"
)

func TestGetConfigArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "long flag",
			args: []string{"-m", "dev", "-config", "/tmp/c.json"},
			want: []string{"-config", "/tmp/c.json"},
		},
		{
			name: "short flag",
			args: []string{"-c", "/tmp/c.json"},
			want: []string{"-c", "/tmp/c.json"},
		},
		{
			name: "missing path",
			args: []string{"-config"},
			want: []string{"-config"},
		},
		{
			name: "no config flag",
			args: []string{"-m", "dev"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetConfigArgs(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetConfigArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
