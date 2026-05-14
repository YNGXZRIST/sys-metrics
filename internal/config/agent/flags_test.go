package agent

import (
	"reflect"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config"
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	t.Helper()

	for _, k := range []string{
		"ADDRESS",
		"MODE",
		"KEY",
		"POLL_INTERVAL",
		"REPORT_INTERVAL",
		"RATE_LIMIT",
		"CRYPTO_KEY",
		"CONFIG",
	} {
		t.Setenv(k, "")
	}
}

func Test_parseArgs(t *testing.T) {
	tests := []struct {
		want    *Options
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "valid args",
			args: []string{
				"-a=127.0.0.1:1234",
				"-r=4",
				"-p=5",
				"-m=development",
				"-crypto-key=/tmp/none.pem",
			},
			want: &Options{
				ServerAddress:  "127.0.0.1:1234",
				Host:           "127.0.0.1",
				Port:           "1234",
				ReportInterval: 4 * time.Second,
				PollInterval:   5 * time.Second,
				ReportSec:      4,
				PollSec:        5,
				Mode:           common.TypeModeDevelopment,
				HashKey:        "",
				RateLimit:      1,
				CryptoKeyPath:  "/tmp/none.pem",
			},
		},
		{
			name: "valid args without crypto",
			args: []string{
				"-a=127.0.0.1:1234",
				"-r=4",
				"-p=5",
				"-m=development",
			},
			want: &Options{
				ServerAddress:  "127.0.0.1:1234",
				Host:           "127.0.0.1",
				Port:           "1234",
				ReportInterval: 4 * time.Second,
				PollInterval:   5 * time.Second,
				ReportSec:      4,
				PollSec:        5,
				Mode:           common.TypeModeDevelopment,
				HashKey:        "",
				RateLimit:      1,
			},
		},
		{
			name: "empty args",
			args: []string{},
			want: &Options{
				ServerAddress:  "localhost:8080",
				Host:           "localhost",
				Port:           "8080",
				ReportInterval: 10 * time.Second,
				PollInterval:   2 * time.Second,
				ReportSec:      10,
				PollSec:        2,
				Mode:           common.TypeModeDefault,
				HashKey:        "",
				RateLimit:      1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)

			opt := new(Options)

			err := opt.parseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"parseArgs() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			err = applyDefaults(opt)
			if err != nil {
				t.Errorf("applyDefaults() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(opt, tt.want) {
				t.Errorf("parseArgs() got = %+v, want %+v", opt, tt.want)
			}
		})
	}
}

func TestOptions_ParseAndSetHostPort(t *testing.T) {
	type fields struct {
		ServerAddress  string
		Host           string
		Port           string
		PollInterval   time.Duration
		ReportInterval time.Duration
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid address",
			fields: fields{
				ServerAddress: "localhost:8080",
				Host:          "localhost",
				Port:          "8080",
			},
			wantErr: false,
		},
		{
			name: "invalid address",
			fields: fields{
				ServerAddress: "localhost",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Options{
				ServerAddress:  tt.fields.ServerAddress,
				Host:           tt.fields.Host,
				Port:           tt.fields.Port,
				PollInterval:   tt.fields.PollInterval,
				ReportInterval: tt.fields.ReportInterval,
			}

			err := config.ParseAndSetHostPort(
				opt.ServerAddress,
				opt,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ParseAndSetHostPort() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func Test_newOption(t *testing.T) {
	tests := []struct {
		want    *Options
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "valid args",
			args: []string{
				"-a=localhost:9090",
				"-r=15",
				"-p=5",
				"-m=development",
			},
			want: &Options{
				ServerAddress:  "localhost:9090",
				Host:           "localhost",
				Port:           "9090",
				ReportInterval: 15 * time.Second,
				PollInterval:   5 * time.Second,
				ReportSec:      15,
				PollSec:        5,
				Mode:           common.TypeModeDevelopment,
				HashKey:        "",
				RateLimit:      1,
			},
		},
		{
			name: "invalid address",
			args: []string{
				"-a=invalid_address",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)

			got, err := NewOption(tt.args)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"NewOption() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf(
					"NewOption() got = %+v, want %+v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestOptions_parseEnv(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantAddress string
		wantHost    string
		wantPort    string
		wantPoll    time.Duration
		wantReport  time.Duration
		wantErr     bool
	}{
		{
			name: "valid env with all variables",
			envVars: map[string]string{
				"ADDRESS":         "localhost:8080",
				"POLL_INTERVAL":   "5",
				"REPORT_INTERVAL": "10",
			},
			wantAddress: "localhost:8080",
			wantHost:    "localhost",
			wantPort:    "8080",
			wantPoll:    5 * time.Second,
			wantReport:  10 * time.Second,
		},
		{
			name: "valid env with custom address",
			envVars: map[string]string{
				"ADDRESS":         "127.0.0.1:9090",
				"POLL_INTERVAL":   "3",
				"REPORT_INTERVAL": "15",
			},
			wantAddress: "127.0.0.1:9090",
			wantHost:    "127.0.0.1",
			wantPort:    "9090",
			wantPoll:    3 * time.Second,
			wantReport:  15 * time.Second,
		},
		{
			name:        "empty env variables",
			envVars:     map[string]string{},
			wantAddress: "",
			wantHost:    "",
			wantPort:    "",
			wantPoll:    0,
			wantReport:  0,
		},
		{
			name: "only intervals set",
			envVars: map[string]string{
				"POLL_INTERVAL":   "7",
				"REPORT_INTERVAL": "20",
			},
			wantPoll:   7 * time.Second,
			wantReport: 20 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)

			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			opt := &Options{}

			err := opt.parseEnv()

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"parseEnv() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if opt.ServerAddress != tt.wantAddress {
				t.Errorf(
					"ServerAddress = %v, want %v",
					opt.ServerAddress,
					tt.wantAddress,
				)
			}

			if opt.Host != tt.wantHost {
				t.Errorf(
					"Host = %v, want %v",
					opt.Host,
					tt.wantHost,
				)
			}

			if opt.Port != tt.wantPort {
				t.Errorf(
					"Port = %v, want %v",
					opt.Port,
					tt.wantPort,
				)
			}

			if opt.PollInterval != tt.wantPoll {
				t.Errorf(
					"PollInterval = %v, want %v",
					opt.PollInterval,
					tt.wantPoll,
				)
			}

			if opt.ReportInterval != tt.wantReport {
				t.Errorf(
					"ReportInterval = %v, want %v",
					opt.ReportInterval,
					tt.wantReport,
				)
			}
		})
	}
}

func TestNewOption_developmentDefaults(t *testing.T) {
	clearEnv(t)

	got, err := NewOption([]string{
		"-m",
		common.TypeModeDevelopment,
	})

	if err != nil {
		t.Fatal(err)
	}

	if got.Mode != common.TypeModeDevelopment {
		t.Fatalf("mode %q", got.Mode)
	}

	if got.Host != "localhost" || got.Port != "8080" {
		t.Fatalf("addr %s:%s", got.Host, got.Port)
	}
}

func TestNewOption_withRateLimitEnv(t *testing.T) {
	clearEnv(t)

	t.Setenv("RATE_LIMIT", "4")

	got, err := NewOption([]string{
		"-m",
		common.TypeModeDevelopment,
	})

	if err != nil {
		t.Fatal(err)
	}

	if got.RateLimit != 4 {
		t.Fatalf("RateLimit = %d", got.RateLimit)
	}
}
