package migrations

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverPairsAndSortsMigrations(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "000002_second.up.sql")
	writeMigrationFile(t, dir, "000001_first.up.sql")
	writeMigrationFile(t, dir, "000002_second.down.sql")
	writeMigrationFile(t, dir, "000001_first.down.sql")

	migrations, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(migrations) != 2 {
		t.Fatalf("len(migrations) = %d, want 2", len(migrations))
	}
	if migrations[0].Version != 1 || migrations[0].Name != "first" {
		t.Fatalf("first migration = %+v, want version 1 name first", migrations[0])
	}
	if migrations[1].Version != 2 || migrations[1].Name != "second" {
		t.Fatalf("second migration = %+v, want version 2 name second", migrations[1])
	}
}

func TestDiscoverRequiresRollbackFile(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "000001_first.up.sql")

	_, err := Discover(dir)
	if err == nil {
		t.Fatal("Discover() error = nil, want missing down migration error")
	}
}

func writeMigrationFile(t *testing.T, dir, name string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("write migration file: %v", err)
	}
}
