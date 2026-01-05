package config

import (
	"reflect"
	"testing"
)

func TestParseServerAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		want    *ServerAddress
		wantErr bool
	}{
		{
			name: "valid server address",
			args: args{
				address: "test:8081",
			},
			want: &ServerAddress{
				Full: "test:8081",
				Host: "test",
				Port: "8081",
			},
			wantErr: false,
		},
		{
			name: "invalid server address",
			args: args{
				address: "localhost",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty server address",
			args: args{
				address: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseServerAddress(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseServerAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseServerAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}
