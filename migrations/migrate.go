package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed *.sql
var files embed.FS

func Apply(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(8401001); CREATE TABLE IF NOT EXISTS schema_migrations(name text PRIMARY KEY, checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	names, err := fs.Glob(files, "*.sql")
	if err != nil {
		return err
	}
	for _, name := range names {
		body, _ := files.ReadFile(name)
		checksum := fmt.Sprintf("%x", sha256.Sum256(body))
		var previous string
		err = tx.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE name=$1`, name).Scan(&previous)
		if err == nil {
			if previous != checksum {
				return fmt.Errorf("migration checksum mismatch: %s", name)
			}
			continue
		}
		if err != sql.ErrNoRows {
			return err
		}
		if _, err = tx.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations(name,checksum) VALUES($1,$2)`, strings.TrimSpace(name), checksum); err != nil {
			return err
		}
	}
	return tx.Commit()
}
