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
		} {
			if _, err := db.ExecContext(ctx, q); err != nil {
				return err
			}
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS notification_sent_history`)
		return err
	})
}
