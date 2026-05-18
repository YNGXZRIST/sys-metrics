package migrations

import (
	"strings"
	"testing"
)

func TestMigrate_emptyDSN(t *testing.T) {
	err := Migrate("")
	if err == nil || !strings.Contains(err.Error(), "not set") {
		t.Fatalf("Migrate: %v", err)
	}
}
