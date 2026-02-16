package middleware

import (
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"testing"
)

func TestWithDBContext(t *testing.T) {
	tests := []struct {
		name      string
		conn      *db.DB
		wantInCtx bool
	}{
		{
			name:      "nil conn does not set context value",
			conn:      nil,
			wantInCtx: false,
		},
		{
			name:      "non-nil conn sets DB in context",
			conn:      &db.DB{},
			wantInCtx: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.Context().Value(common.ContextDBKey)
			})

			handler := WithDBContext(tt.conn)(next)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if tt.wantInCtx && got == nil {
				t.Error("expected DB in context, got nil")
			}
			if !tt.wantInCtx && got != nil {
				t.Errorf("expected no DB in context, got %v", got)
			}
		})
	}
}
