package migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	Migrations.MustRegister(renameIntervalUp, renameIntervalDown)
}

// renameIntervalUp renames the 'interval' column to 'check_interval' in monitors.
// No-op if check_interval already exists (fresh installs skip this).
func renameIntervalUp(ctx context.Context, db *bun.DB) error {
	var sql string
	if db.Dialect().Name() == dialect.MySQL {
		sql = "ALTER TABLE monitors CHANGE COLUMN `interval` check_interval INTEGER NOT NULL"
	} else {
		// SQLite and PostgreSQL both support RENAME COLUMN
		sql = `ALTER TABLE monitors RENAME COLUMN "interval" TO check_interval`
	}

	if _, err := db.ExecContext(ctx, sql); err != nil {
		s := err.Error()
		// Column already named check_interval — fresh install or already migrated.
		if strings.Contains(s, "no such column") ||
			strings.Contains(s, "Unknown column") ||
			strings.Contains(s, "does not exist") ||
			strings.Contains(s, "no column named") {
			return nil
		}
		return err
	}
	return nil
}

func renameIntervalDown(ctx context.Context, db *bun.DB) error {
	var sql string
	if db.Dialect().Name() == dialect.MySQL {
		sql = "ALTER TABLE monitors CHANGE COLUMN check_interval `interval` INTEGER NOT NULL"
	} else {
		sql = `ALTER TABLE monitors RENAME COLUMN check_interval TO "interval"`
	}
	_, err := db.ExecContext(ctx, sql)
	return err
}
