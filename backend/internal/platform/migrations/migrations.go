package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version bigint PRIMARY KEY,
  name text NOT NULL,
  applied_at timestamptz NOT NULL DEFAULT now()
)`

type Migration struct {
	Version  int64
	Name     string
	UpPath   string
	DownPath string
}

func ApplyUp(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if pool == nil {
		return fmt.Errorf("database pool is not configured")
	}

	discovered, err := Discover(dir)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)

	if _, err := tx.Exec(ctx, schemaTableSQL); err != nil {
		return err
	}

	applied, err := appliedVersions(ctx, tx)
	if err != nil {
		return err
	}

	for _, migration := range discovered {
		if applied[migration.Version] {
			continue
		}
		sqlBytes, err := os.ReadFile(migration.UpPath)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.Name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, migration.Version, migration.Name); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func ApplyDown(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if pool == nil {
		return fmt.Errorf("database pool is not configured")
	}

	discovered, err := Discover(dir)
	if err != nil {
		return err
	}
	byVersion := make(map[int64]Migration, len(discovered))
	for _, migration := range discovered {
		byVersion[migration.Version] = migration
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)

	if _, err := tx.Exec(ctx, schemaTableSQL); err != nil {
		return err
	}

	var version int64
	err = tx.QueryRow(ctx, `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version)
	if err == pgx.ErrNoRows {
		return tx.Commit(ctx)
	}
	if err != nil {
		return err
	}

	migration, ok := byVersion[version]
	if !ok {
		return fmt.Errorf("applied migration %d has no local files", version)
	}

	sqlBytes, err := os.ReadFile(migration.DownPath)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("rollback migration %s: %w", migration.Name, err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, version); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func Discover(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	byVersion := map[int64]*Migration{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		version, title, direction, ok := parseMigrationFilename(name)
		if !ok {
			return nil, fmt.Errorf("invalid migration filename %q", name)
		}

		migration := byVersion[version]
		if migration == nil {
			migration = &Migration{Version: version, Name: title}
			byVersion[version] = migration
		}
		if migration.Name != title {
			return nil, fmt.Errorf("migration version %d has mismatched names", version)
		}

		path := filepath.Join(dir, name)
		switch direction {
		case "up":
			migration.UpPath = path
		case "down":
			migration.DownPath = path
		default:
			return nil, fmt.Errorf("invalid migration direction %q", direction)
		}
	}

	migrations := make([]Migration, 0, len(byVersion))
	for _, migration := range byVersion {
		if migration.UpPath == "" || migration.DownPath == "" {
			return nil, fmt.Errorf("migration %d_%s must include up and down files", migration.Version, migration.Name)
		}
		migrations = append(migrations, *migration)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigrationFilename(filename string) (int64, string, string, bool) {
	trimmed := strings.TrimSuffix(filename, ".sql")
	parts := strings.Split(trimmed, ".")
	if len(parts) != 2 {
		return 0, "", "", false
	}

	nameParts := strings.SplitN(parts[0], "_", 2)
	if len(nameParts) != 2 {
		return 0, "", "", false
	}

	version, err := strconv.ParseInt(nameParts[0], 10, 64)
	if err != nil || version < 1 {
		return 0, "", "", false
	}

	direction := parts[1]
	if direction != "up" && direction != "down" {
		return 0, "", "", false
	}

	return version, nameParts[1], direction, true
}

func appliedVersions(ctx context.Context, tx pgx.Tx) (map[int64]bool, error) {
	rows, err := tx.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := map[int64]bool{}
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return versions, nil
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}
