package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

func runEmbeddedSQLMigrations(db *DB) error {
	if db == nil || db.conn == nil {
		return nil
	}
	const dir = "migrations"
	entries, err := embeddedMigrations.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		version := strings.TrimSuffix(name, ".sql")
		var dummy int
		err := db.conn.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = $1`, version).Scan(&dummy)
		if err == nil {
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("schema_migrations check %q: %w", version, err)
		}
		path := filepath.Join(dir, name)
		body, err := embeddedMigrations.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", path, err)
		}
		tx, err := db.conn.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %q: %w", version, err)
		}
		if _, err := tx.Exec(string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute migration %q: %w", version, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %q: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %q: %w", version, err)
		}
		if db.logger != nil {
			db.logger.Info("applied sql migration", zap.String("version", version))
		}
	}
	return nil
}
