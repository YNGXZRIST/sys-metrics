package utils

import (
	"bytes"
	"io"
	"os"
	"sys-metrics/internal/common"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = orig

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func TestPrintBuildInfo(t *testing.T) {
	type args struct {
		version string
		date    string
		commit  string
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "all empty -> N/A",
			args: args{},
			expected: "Build version: " + common.NotApplicable + "\n" +
				"Build date: " + common.NotApplicable + "\n" +
				"Build commit: " + common.NotApplicable + "\n",
		},
		{
			name: "only version set",
			args: args{version: "v1.0.1"},
			expected: "Build version: v1.0.1\n" +
				"Build date: " + common.NotApplicable + "\n" +
				"Build commit: " + common.NotApplicable + "\n",
		},
		{
			name: "only date set",
			args: args{date: "2026-05-06"},
			expected: "Build version: " + common.NotApplicable + "\n" +
				"Build date: 2026-05-06\n" +
				"Build commit: " + common.NotApplicable + "\n",
		},
		{
			name: "only commit set",
			args: args{commit: "deadbeef"},
			expected: "Build version: " + common.NotApplicable + "\n" +
				"Build date: " + common.NotApplicable + "\n" +
				"Build commit: deadbeef\n",
		},
		{
			name: "all set",
			args: args{version: "v1.0.1", date: "2026-05-06", commit: "deadbeef"},
			expected: "Build version: v1.0.1\n" +
				"Build date: 2026-05-06\n" +
				"Build commit: deadbeef\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStdout(t, func() {
				PrintBuildInfo(tt.args.version, tt.args.date, tt.args.commit)
			})
			if got != tt.expected {
				t.Fatalf("unexpected output:\n--- got ---\n%s\n--- want ---\n%s", got, tt.expected)
			}
		})
	}
}
