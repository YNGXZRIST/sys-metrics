package migrations

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedMigrationsFS(t *testing.T) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected embedded SQL migrations")
	}
	hasSQL := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			hasSQL = true
			data, err := FS.ReadFile(e.Name())
			if err != nil || len(data) == 0 {
				t.Fatalf("file %q: err=%v len=%d", e.Name(), err, len(data))
			}
		}
	}
	if !hasSQL {
		t.Fatal("no .sql files in embed FS")
	}
}

func TestMigrate_invalidDSN(t *testing.T) {
	err := Migrate("postgres://127.0.0.1:1/nodb?sslmode=disable&connect_timeout=1")
	if err == nil {
		t.Fatal("expected migrate error")
	}
}
