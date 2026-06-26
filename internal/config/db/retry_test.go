package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestRunWithRetry_success(t *testing.T) {
	got, err := runWithRetry(context.Background(), func() (int, error) {
		return 42, nil
	})
	if err != nil || got != 42 {
		t.Fatalf("got=%d err=%v", got, err)
	}
}

func TestRunWithRetry_nonRetriable(t *testing.T) {
	_, err := runWithRetry(context.Background(), func() (int, error) {
		return 0, errors.New("fatal")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunWithRetry_contextCanceled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := runWithRetry(ctx, func() (int, error) {
		return 0, &pgconn.PgError{Code: "40001"}
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDB_BeginTx_invalidDSN(t *testing.T) {
	conn, err := NewConn(NewCfg(&Config{DNS: "postgres://invalid:5432/nodb?sslmode=disable"}))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.BeginTx(context.Background(), nil)
	if err == nil {
		t.Fatal("expected begin tx error")
	}
}
