package db

import (
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type migration struct {
	version int64
	name    string
	sql     string
	hash    string
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint PRIMARY KEY,
			name text NOT NULL,
			checksum text NOT NULL,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	for _, m := range migrations {
		var appliedHash string
		err := pool.QueryRow(ctx, `SELECT checksum FROM schema_migrations WHERE version = $1`, m.version).Scan(&appliedHash)
		if err == nil {
			if appliedHash != m.hash {
				return fmt.Errorf("migration %d (%s) checksum mismatch", m.version, m.name)
			}
			continue
		}
		if !isNoRows(err) {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.version, err)
		}
		if _, err := tx.Exec(ctx, m.sql); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %d (%s): %w", m.version, m.name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, name, checksum) VALUES ($1, $2, $3)`,
			m.version, m.name, m.hash); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}
	}
	return nil
}

func loadMigrations() ([]migration, error) {
	entries, err := fs.Glob(migrationFS, "migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(entries)
	result := make([]migration, 0, len(entries))
	for _, file := range entries {
		base := path.Base(file)
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid migration filename %q", base)
		}
		version, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version %q: %w", base, err)
		}
		data, err := fs.ReadFile(migrationFS, file)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", base, err)
		}
		hash := fmt.Sprintf("%x", sha256.Sum256(data))
		result = append(result, migration{version: version, name: base, sql: string(data), hash: hash})
	}
	return result, nil
}

func isNoRows(err error) bool { return err == pgx.ErrNoRows }
