package context

import (
	"context"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"testing"

	"go.uber.org/zap"
)

func TestDBFromContext(t *testing.T) {
	// Use nil-embedded *db.DB for test (no real connection needed)
	dbVal := &db.DB{}

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
		wantDB  *db.DB
	}{
		{
			name:    "DB in context returns connection",
			ctx:     context.WithValue(context.Background(), common.ContextDBKey, dbVal),
			wantErr: false,
			wantDB:  dbVal,
		},
		{
			name:    "empty context returns error",
			ctx:     context.Background(),
			wantErr: true,
			wantDB:  nil,
		},
		{
			name:    "wrong type in context returns error",
			ctx:     context.WithValue(context.Background(), common.ContextDBKey, "not-a-db"),
			wantErr: true,
			wantDB:  nil,
		},
		{
			name:    "nil value for key returns error",
			ctx:     context.WithValue(context.Background(), common.ContextDBKey, nil),
			wantErr: true,
			wantDB:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DBFromContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err != nil && err.Error() != "db context not found in context" {
					t.Errorf("DBFromContext() error = %q", err.Error())
				}
				return
			}
			if got != tt.wantDB {
				t.Errorf("DBFromContext() = %v, want %v", got, tt.wantDB)
			}
		})
	}
}

func TestLoggerFromContext(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		ctx     context.Context
		wantNop bool
		want    *zap.Logger
	}{
		{
			name:    "logger in context returns it",
			ctx:     context.WithValue(context.Background(), common.ContextLoggerKey, logger),
			wantNop: false,
			want:    logger,
		},
		{
			name:    "empty context returns nop logger",
			ctx:     context.Background(),
			wantNop: true,
			want:    nil,
		},
		{
			name:    "wrong type in context returns nop logger",
			ctx:     context.WithValue(context.Background(), common.ContextLoggerKey, "not-a-logger"),
			wantNop: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LoggerFromContext(tt.ctx)
			if got == nil {
				t.Fatal("LoggerFromContext() never returns nil")
			}
			if tt.wantNop {
				// Nop logger is returned; we can't compare by pointer, just check we got something
				return
			}
			if got != tt.want {
				t.Errorf("LoggerFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
