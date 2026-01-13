package main

import (
	"reflect"
	"testing"
	"time"
)

func Test_parseArgs(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		args    args
		want    *Options
		wantErr bool
	}{
		{
			name: "valid args",
			args: args{
				args: []string{
					"-a=127.0.0.1:1234",
					"-r=4",
					"-p=5",
				},
			},
			want: &Options{
				serverAddress:  "127.0.0.1:1234",
				host:           "127.0.0.1",
				port:           "1234",
				reportInterval: 4 * time.Second,
				pollInterval:   5 * time.Second,
			},
		},
		{
			name: "Empty args",
			args: args{
				args: []string{},
			},
			want: &Options{
				serverAddress:  "localhost:8080",
				host:           "localhost",
				port:           "8080",
				reportInterval: 10 * time.Second,
				pollInterval:   2 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseArgs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
