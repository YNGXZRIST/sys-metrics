package postgres

import (
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/config/server"
	"testing"
)

func TestConfig_NeedSync(t *testing.T) {
	opt := &server.Options{
		DNS: "test",
	}
	conn, err := db.NewConn(db.NewCfg(opt))
	if err != nil {
		t.Fatal("Failed to create new connection")
	}
	cfg := NewConfig(conn)
	if cfg.NeedSync() {
		t.Fatal("Need Sync is true, want false")
	}
	cfg.initialized = true
	if !cfg.NeedSync() {
		t.Fatal("Need Sync is false, want true")
	}

}

func TestNewConfig(t *testing.T) {
	type args struct {
		dbConn *db.DB
	}
	tests := []struct {
		args args
		name string
	}{
		{
			name: "nil dbConn",
			args: args{dbConn: nil},
		},
		{
			name: "non-nil dbConn",
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "non-nil dbConn" {
				conn, err := db.NewConn(db.NewCfg(&server.Options{DNS: "test"}))
				if err != nil {
					t.Skipf("skip: no db: %v", err)
				}
				defer conn.Close()
				tt.args.dbConn = conn
			}
			got := NewConfig(tt.args.dbConn)
			if got == nil {
				t.Fatal("NewConfig() returned nil")
			}
			if got.conn != tt.args.dbConn {
				t.Errorf("NewConfig().conn = %v, want %v", got.conn, tt.args.dbConn)
			}
			if got.handler == nil {
				t.Error("NewConfig().handler should not be nil")
			}
			if got.initialized {
				t.Error("NewConfig().initialized should be false")
			}
			if tt.args.dbConn != nil {
				h, ok := got.handler.(*Handler)
				if !ok {
					t.Errorf("NewConfig().handler type = %T, want *Handler", got.handler)
				} else if h.dbConn != tt.args.dbConn {
					t.Error("NewConfig().handler.dbConn != conn")
				}
			}
		})
	}
}
