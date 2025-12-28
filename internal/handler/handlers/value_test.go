package handlers

import (
	"net/http"
	"testing"
)

func TestValueHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ValueHandler(tt.args.w, tt.args.r)
		})
	}
}
