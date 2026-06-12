package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		idCol := "SERIAL PRIMARY KEY"
		if db.Dialect().Name() == dialect.MySQL {
			idCol = "INT NOT NULL AUTO_INCREMENT PRIMARY KEY"
		}
		for _, q := range []string{
			fmt.Sprintf(`CREATE TABLE monitor_tls_info (
    id %s,
    monitor_id VARCHAR(255) NOT NULL UNIQUE,
    info_json TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`, idCol),
			`CREATE INDEX idx_monitor_tls_info_monitor_id ON monitor_tls_info(monitor_id)`,
			`CREATE INDEX idx_monitor_tls_info_updated_at ON monitor_tls_info(updated_at)`,
		} {
			if _, err := db.ExecContext(ctx, q); err != nil {
				return err
			}
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS monitor_tls_info`)
		return err
	})
}
