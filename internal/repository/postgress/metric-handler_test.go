//go:build integration
// +build integration

package postgress

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"reflect"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/config/server"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/migrations"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var testDB *db.DB
var testDSN string

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		os.Exit(1)
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
		os.Exit(1)
	}
	_ = resource.Expire(30)

	hostPort := resource.GetPort("5432/tcp")
	dsn := "postgres://postgres:postgres@localhost:" + hostPort + "/postgres?sslmode=disable"

	if err = pool.Retry(func() error {
		conn, retryErr := sql.Open("pgx", dsn)
		if retryErr != nil {
			return retryErr
		}
		defer conn.Close()
		return conn.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		os.Exit(1)
	}
	if err = migrations.Migrate(dsn); err != nil {
		_ = pool.Purge(resource)
		os.Exit(1)
	}

	testDSN = dsn
	testDB, err = db.NewConn(db.NewCfg(&server.Options{DNS: dsn}))
	if err != nil {
		_ = pool.Purge(resource)
		os.Exit(1)
	}

	code := m.Run()
	_ = testDB.Close()
	_ = pool.Purge(resource)
	os.Exit(code)
}

func getTestDB(t *testing.T) *db.DB {
	t.Helper()
	if testDB == nil {
		t.Skip("Test DB not available (e.g. Docker not running)")
	}
	return testDB
}

func TestHandler_Read(t *testing.T) {
	conn := getTestDB(t)
	cfg := NewConfig(conn)
	_ = cfg
	gauge := metrics.NewGauge(common.Alloc)
	gauge.SetValue(1)
	counter := metrics.NewCounter(common.PollCount)
	counter.SetValue(2)
	_, err := conn.Exec("INSERT INTO metrics (name, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5), ($6, $7, $8, $9, $10)",
		gauge.ID, gauge.MType, gauge.Delta, gauge.Value, gauge.Hash, counter.ID, counter.MType, counter.Delta, counter.Value, counter.Hash)
	if err != nil {
		t.Fatal("cannot insert metrics", err)
	}
	read, err := cfg.handler.Read(context.TODO())
	if err != nil {
		t.Fatal("cannot read metrics from database", err)
	}
	for _, m := range read {
		switch m.MType {
		case common.Gauge:
			if !reflect.DeepEqual(m, gauge.Metrics) {
				t.Fatalf("metrics gauge are not equal: got %+v, want %+v", m, gauge.Metrics)
			}
		case common.Counter:
			if !reflect.DeepEqual(m, counter.Metrics) {
				t.Fatalf("metrics counter are not equal: got %+v, want %+v", m, counter.Metrics)
			}
		}
	}
}

func TestHandler_Upsert(t *testing.T) {
	conn := getTestDB(t)
	cfg := NewConfig(conn)

	gauge := metrics.NewGauge(common.TypeModeTest)
	gauge.SetValue(1)
	metric := &metrics.Metrics{ID: gauge.ID, MType: gauge.MType, Delta: gauge.Delta, Value: gauge.Value, Hash: gauge.Hash}
	err := cfg.handler.Upsert(context.TODO(), metric)
	if err != nil {
		t.Fatalf("Could not upsert metric: %s", err)
	}
	var m metrics.Metrics
	var rowID int64
	var hash sql.NullString
	err = cfg.conn.QueryRow("SELECT id, name, mtype, delta, value, hash FROM metrics WHERE name = $1", metric.ID).Scan(&rowID, &m.ID, &m.MType, &m.Delta, &m.Value, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Could not find upsert metric: %s", err)
	} else if err != nil {
		t.Fatalf("unknown find error: %s", err)
	}
	if hash.Valid {
		m.Hash = hash.String
	}
	if !reflect.DeepEqual(m, *metric) {
		t.Fatalf("metrics are not equal: got %+v, want %+v", m, *metric)
	}

}

func TestHandler_Write(t *testing.T) {
	conn := getTestDB(t)
	cfg := NewConfig(conn)

	gauge := metrics.NewGauge(common.TypeModeTest)
	gauge.SetValue(2)
	metric := &metrics.Metrics{ID: gauge.ID, MType: gauge.MType, Delta: gauge.Delta, Value: gauge.Value, Hash: gauge.Hash}
	err := cfg.handler.Write(context.TODO(), metric)
	if err != nil {
		t.Fatalf("Could not write metric: %s", err)
	}

	var m metrics.Metrics
	var rowID int64
	var hash sql.NullString
	err = cfg.conn.QueryRow("SELECT id, name, mtype, delta, value, hash FROM metrics WHERE name = $1", metric.ID).Scan(&rowID, &m.ID, &m.MType, &m.Delta, &m.Value, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Could not find written metric: %s", err)
	} else if err != nil {
		t.Fatalf("unknown find error: %s", err)
	}
	if hash.Valid {
		m.Hash = hash.String
	}
	if !reflect.DeepEqual(m, *metric) {
		t.Fatalf("metrics are not equal: got %+v, want %+v", m, *metric)
	}
}

func TestHandler_WriteBatch(t *testing.T) {
	conn := getTestDB(t)
	cfg := NewConfig(conn)
	gauge := metrics.NewGauge(common.Alloc)
	gauge.SetValue(1)
	counter := metrics.NewCounter(common.PollCount)
	counter.SetValue(2)
	metricsArr := []metrics.Metrics{gauge.Metrics, counter.Metrics}
	err := cfg.handler.WriteBatch(context.TODO(), metricsArr)
	if err != nil {
		t.Fatalf("Could not write metrics: %s", err)
	}
	read, err := cfg.handler.Read(context.TODO())
	if err != nil {
		t.Fatalf("Could not read after WriteBatch: %s", err)
	}
	if len(read) < 2 {
		t.Fatalf("expected at least 2 metrics, got %d", len(read))
	}
	gotIDs := make(map[string]metrics.Metrics)
	for _, m := range read {
		gotIDs[m.ID] = m
	}
	for _, want := range metricsArr {
		got, ok := gotIDs[want.ID]
		if !ok {
			t.Fatalf("metric %q not found after WriteBatch", want.ID)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("metric %q: got %+v, want %+v", want.ID, got, want)
		}
	}
}

func TestHandler_chunkSelect(t *testing.T) {
	conn := getTestDB(t)
	h := &Handler{dbConn: conn}
	ctx := context.TODO()

	t.Run("empty when lastID beyond data", func(t *testing.T) {
		lastID := int64(999999999)
		got, got1, err := h.chunkSelect(ctx, lastID)
		if err != nil {
			t.Fatalf("chunkSelect() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("chunkSelect() got %d rows, want 0", len(got))
		}
		if got1 != lastID {
			t.Errorf("chunkSelect() got1 = %v, want %v (unchanged when no rows)", got1, lastID)
		}
	})

	t.Run("returns seeded data", func(t *testing.T) {
		_, err := conn.ExecContext(ctx, "TRUNCATE TABLE metrics RESTART IDENTITY")
		if err != nil {
			t.Fatalf("truncate: %v", err)
		}
		gauge := metrics.NewGauge(common.Alloc)
		gauge.SetValue(1)
		counter := metrics.NewCounter(common.PollCount)
		counter.SetValue(2)
		cfg := NewConfig(conn)
		err = cfg.handler.WriteBatch(ctx, []metrics.Metrics{gauge.Metrics, counter.Metrics})
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
		got, got1, err := h.chunkSelect(ctx, 0)
		if err != nil {
			t.Fatalf("chunkSelect() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("chunkSelect() got %d rows, want 2", len(got))
		}
		if got1 < 1 {
			t.Errorf("chunkSelect() got1 = %v, want positive (last row id)", got1)
		}
		byID := make(map[string]metrics.Metrics)
		for _, m := range got {
			byID[m.ID] = m
		}
		if !reflect.DeepEqual(byID[gauge.ID], gauge.Metrics) {
			t.Errorf("gauge: got %+v, want %+v", byID[gauge.ID], gauge.Metrics)
		}
		if !reflect.DeepEqual(byID[counter.ID], counter.Metrics) {
			t.Errorf("counter: got %+v, want %+v", byID[counter.ID], counter.Metrics)
		}
	})
}

func TestNewHandler(t *testing.T) {
	t.Run("nil dbConn", func(t *testing.T) {
		got := NewHandler(nil)
		if got == nil {
			t.Fatal("NewHandler(nil) returned nil")
		}
		if got.dbConn != nil {
			t.Errorf("NewHandler(nil).dbConn = %v, want nil", got.dbConn)
		}
	})

	t.Run("non-nil dbConn", func(t *testing.T) {
		conn := getTestDB(t)
		got := NewHandler(conn)
		if got == nil {
			t.Fatal("NewHandler(Conn) returned nil")
		}
		if got.dbConn != conn {
			t.Error("NewHandler(Conn).dbConn != conn")
		}
	})
}
