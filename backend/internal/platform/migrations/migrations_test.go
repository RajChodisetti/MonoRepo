package migrations

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestValidateAppliedMigrationsRequiresExactVersionNameParity(t *testing.T) {
	discovered := []Migration{
		{Version: 39, Name: "company_consultation_slot_overrides"},
		{Version: 40, Name: "keep_supported_demo_templates"},
	}

	tests := []struct {
		name        string
		applied     map[int64]string
		wantErrPart string
	}{
		{
			name: "matching lineage",
			applied: map[int64]string{
				39: "company_consultation_slot_overrides",
				40: "keep_supported_demo_templates",
			},
		},
		{
			name: "same version different branch name",
			applied: map[int64]string{
				39: "company_consultation_slot_overrides",
				40: "different_branch_migration",
			},
			wantErrPart: `migration lineage mismatch at version 40: database has "different_branch_migration", local files have "keep_supported_demo_templates"`,
		},
		{
			name: "applied version missing locally",
			applied: map[int64]string{
				41: "unknown_branch_migration",
			},
			wantErrPart: "applied migration 41_unknown_branch_migration has no matching local files",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateAppliedMigrations(discovered, test.applied)
			if test.wantErrPart == "" {
				if err != nil {
					t.Fatalf("validateAppliedMigrations() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErrPart) {
				t.Fatalf("validateAppliedMigrations() error = %v, want containing %q", err, test.wantErrPart)
			}
		})
	}
}

func TestRepositoryMigrationsIncludeConsultationSlotOverrides(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() did not return this test file")
	}
	dir := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "migrations")

	migrations, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover(repository migrations) error = %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("Discover(repository migrations) returned no migrations")
	}
	for _, migration := range migrations {
		if migration.Version == 39 && migration.Name == "company_consultation_slot_overrides" {
			return
		}
	}
	t.Fatal("repository migrations do not include 39_company_consultation_slot_overrides")
}

func TestRepositoryLatestMigrationAddsConsultationOverlapAndCalendarRevisions(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() did not return this test file")
	}
	dir := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "migrations")

	migrations, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover(repository migrations) error = %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("Discover(repository migrations) returned no migrations")
	}
	latest := migrations[len(migrations)-1]
	if latest.Version != 41 || latest.Name != "consultation_overlap_and_calendar_revisions" {
		t.Fatalf("latest migration = %d_%s, want 41_consultation_overlap_and_calendar_revisions", latest.Version, latest.Name)
	}
	upSQL, err := os.ReadFile(latest.UpPath)
	if err != nil {
		t.Fatalf("read latest up migration: %v", err)
	}
	for _, required := range []string{
		"company_consultations_valid_interval",
		"CHECK (slot_end > slot_start)",
		"EXCLUDE USING gist",
		"tstzrange(slot_start, slot_end, '[)') WITH &&",
		"WHERE (status = 'confirmed')",
		"CREATE TABLE company_consultation_calendar_months",
	} {
		if !strings.Contains(string(upSQL), required) {
			t.Fatalf("latest up migration missing %q", required)
		}
	}
	downSQL, err := os.ReadFile(latest.DownPath)
	if err != nil {
		t.Fatalf("read latest down migration: %v", err)
	}
	for _, required := range []string{
		"DROP CONSTRAINT IF EXISTS company_consultations_valid_interval",
		"DROP CONSTRAINT IF EXISTS company_consultations_confirmed_no_overlap",
		"DROP TABLE IF EXISTS company_consultation_calendar_months",
	} {
		if !strings.Contains(string(downSQL), required) {
			t.Fatalf("latest down migration missing %q", required)
		}
	}
}

func writeMigrationFile(t *testing.T, dir, name string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("write migration file: %v", err)
	}
}
