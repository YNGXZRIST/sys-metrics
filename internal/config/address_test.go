package config

import (
	"reflect"
	"sys-metrics/internal/common"
	"testing"
)

func TestParseServerAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		want    *ServerAddress
		name    string
		args    args
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

type hostPortRecorder struct {
	host, port string
}

func (h *hostPortRecorder) SetHostPort(host, port string) {
	h.host, h.port = host, port
}

func TestParseAndSetHostPort(t *testing.T) {
	var rec hostPortRecorder
	if err := ParseAndSetHostPort("api.example.com:443", &rec); err != nil {
		t.Fatal(err)
	}
	if rec.host != "api.example.com" || rec.port != "443" {
		t.Fatalf("got %q:%q", rec.host, rec.port)
	}
	if err := ParseAndSetHostPort("bad", &rec); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateMode(t *testing.T) {
	if err := ValidateMode(common.TypeModeDevelopment); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMode(common.TypeModeProduction); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMode("staging"); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateReportTransport(t *testing.T) {
	if err := ValidateReportTransport(common.ReportTransportHTTP); err != nil {
		t.Fatal(err)
	}
	if err := ValidateReportTransport(common.ReportTransportGRPC); err != nil {
		t.Fatal(err)
	}
	if err := ValidateReportTransport("kafka"); err == nil {
		t.Fatal("expected error")
	}
}
