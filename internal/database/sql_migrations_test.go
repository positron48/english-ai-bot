package database

import (
	"testing"
)

func TestEmbeddedMigrationsPresent(t *testing.T) {
	entries, err := embeddedMigrations.ReadDir("migrations")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one .sql migration")
	}
}
