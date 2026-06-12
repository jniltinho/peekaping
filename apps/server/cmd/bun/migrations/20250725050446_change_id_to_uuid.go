package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		for _, q := range []string{
			`DROP TABLE IF EXISTS notification_sent_history`,
			`CREATE TABLE notification_sent_history (
    id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    monitor_id VARCHAR(255) NOT NULL,
    days INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(type, monitor_id, days)
)`,
			`CREATE INDEX idx_notification_sent_history_type_monitor ON notification_sent_history(type, monitor_id)`,
			`CREATE INDEX idx_notification_sent_history_created_at ON notification_sent_history(created_at)`,
			`DROP TABLE IF EXISTS monitor_tls_info`,
			`CREATE TABLE monitor_tls_info (
    id UUID PRIMARY KEY,
    monitor_id VARCHAR(255) NOT NULL UNIQUE,
    info_json TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
			`CREATE INDEX idx_monitor_tls_info_monitor_id ON monitor_tls_info(monitor_id)`,
			`CREATE INDEX idx_monitor_tls_info_updated_at ON monitor_tls_info(updated_at)`,
		} {
			if _, err := db.ExecContext(ctx, q); err != nil {
				return err
			}
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		idCol := "SERIAL PRIMARY KEY"
		if db.Dialect().Name() == dialect.MySQL {
			idCol = "INT NOT NULL AUTO_INCREMENT PRIMARY KEY"
		}
		for _, q := range []string{
			`DROP TABLE IF EXISTS notification_sent_history`,
			fmt.Sprintf(`CREATE TABLE notification_sent_history (
    id %s,
    type VARCHAR(50) NOT NULL,
    monitor_id VARCHAR(255) NOT NULL,
    days INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(type, monitor_id, days)
)`, idCol),
			`CREATE INDEX idx_notification_sent_history_type_monitor ON notification_sent_history(type, monitor_id)`,
			`CREATE INDEX idx_notification_sent_history_created_at ON notification_sent_history(created_at)`,
			`DROP TABLE IF EXISTS monitor_tls_info`,
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
	})
}
