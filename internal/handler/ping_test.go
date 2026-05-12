//go test --tags=integration -v -run TestPingHandler_DBSuccess
//go:build integration
// +build integration

package handler

import (
	"net/http"
	"net/http/httptest"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/config/server"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.uber.org/zap"
)

func TestPingHandler_DBSuccess(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	logger := zap.NewNop()

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("Could not start postgres: %s", err)
	}
	err = resource.Expire(30)
	if err != nil {
		t.Fatalf("Could not end resource: %s", err)
	}
	defer pool.Purge(resource)

	hostPort := resource.GetPort("5432/tcp")
	t.Logf("Postgres is available on port: %s", hostPort)
	dsn := "postgres://postgres:postgres@localhost:" + hostPort + "/postgres?sslmode=disable"
	err = pool.Retry(func() error {
		cfg := db.NewCfg(&server.Options{DNS: dsn})
		conn, err := db.NewConn(cfg)
		if err != nil {
			return err
		}
		defer func(conn *db.DB) {
			err := conn.Close()
			if err != nil {
				t.Errorf("failed to close db connection: %v", err)
			}
		}(conn)
		return conn.Ping()
	})
	if err != nil {
		t.Fatalf("Could not connect to database: %s", err)
	}

	opt := &server.Options{
		DNS: dsn,
	}
	cfg := db.NewCfg(opt)
	conn, err := db.NewConn(cfg)
	if err != nil {
		t.Fatalf("failed to create db connection: %v", err)
	}
	defer func(conn *db.DB) {
		err := conn.Close()
		if err != nil {
			t.Errorf("failed to close db connection: %v", err)
		}
	}(conn)

	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	h := NewHandler(conn, nil, nil, logger, nil)
	h.PingHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("PingHandler returned status %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestPingHandler_DBError(t *testing.T) {
	logger := zap.NewNop()
	opt := &server.Options{
		DNS: "postgres://wrong:wrong@localhost:5432/wrong?sslmode=disable",
	}
	cfg := db.NewCfg(opt)
	conn, _ := db.NewConn(cfg)

	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	h := NewHandler(conn, nil, nil, logger, nil)
	h.PingHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("PingHandler returned status %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
func TestPingHandler_EmptyContext(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	h := NewHandler(nil, nil, nil, zap.NewNop(), nil)
	h.PingHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("PingHandler returned status %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
