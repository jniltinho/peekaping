package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		stmts := []string{
			`CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    twofa_secret VARCHAR(64),
    twofa_status BOOLEAN NOT NULL DEFAULT false,
    twofa_last_token VARCHAR(6),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`,
			`CREATE TABLE IF NOT EXISTS proxies (
    id UUID PRIMARY KEY,
    protocol VARCHAR(10) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    auth BOOLEAN NOT NULL DEFAULT false,
    username VARCHAR(255),
    password VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`,
			// 'interval' is reserved in MySQL — renamed to check_interval
			`CREATE TABLE IF NOT EXISTS monitors (
    id UUID PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    name VARCHAR(150) NOT NULL,
    check_interval INTEGER NOT NULL,
    timeout INTEGER NOT NULL,
    max_retries INTEGER NOT NULL,
    retry_interval INTEGER NOT NULL,
    resend_interval INTEGER NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    status INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    config JSON,
    proxy_id UUID,
    push_token VARCHAR(255),
    FOREIGN KEY (proxy_id) REFERENCES proxies(id) ON DELETE SET NULL
)`,
			`CREATE TABLE IF NOT EXISTS status_pages (
    id UUID PRIMARY KEY,
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(255),
    theme VARCHAR(30) NOT NULL DEFAULT 'light',
    published BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    footer_text TEXT,
    auto_refresh_interval INTEGER NOT NULL DEFAULT 300
)`,
			`CREATE TABLE IF NOT EXISTS notification_channels (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    config JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`,
			`CREATE TABLE IF NOT EXISTS maintenances (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    strategy VARCHAR(50) NOT NULL,
    start_date_time VARCHAR(50),
    end_date_time VARCHAR(50),
    start_time VARCHAR(50),
    end_time VARCHAR(50),
    weekdays TEXT,
    days_of_month TEXT,
    interval_day INTEGER,
    cron VARCHAR(255),
    timezone VARCHAR(100),
    duration INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`,
			// 'key' is reserved in MySQL — renamed to setting_key
			`CREATE TABLE IF NOT EXISTS settings (
    setting_key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`,
			// 'time' is reserved in MySQL — renamed to check_time
			`CREATE TABLE IF NOT EXISTS heartbeats (
    id UUID PRIMARY KEY,
    monitor_id UUID NOT NULL,
    status INTEGER NOT NULL,
    msg TEXT,
    ping INTEGER,
    duration INTEGER,
    down_count INTEGER,
    retries INTEGER,
    important BOOLEAN NOT NULL DEFAULT false,
    check_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP,
    notified BOOLEAN NOT NULL DEFAULT false,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
)`,
			// 'timestamp' is reserved in MySQL — renamed to stat_ts
			`CREATE TABLE IF NOT EXISTS stats (
    id UUID PRIMARY KEY,
    monitor_id UUID NOT NULL,
    stat_ts TIMESTAMP NOT NULL,
    ping DOUBLE PRECISION NOT NULL DEFAULT 0,
    ping_min DOUBLE PRECISION NOT NULL DEFAULT 0,
    ping_max DOUBLE PRECISION NOT NULL DEFAULT 0,
    up INTEGER NOT NULL DEFAULT 0,
    down INTEGER NOT NULL DEFAULT 0,
    maintenance INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE,
    UNIQUE(monitor_id, stat_ts)
)`,
			`CREATE TABLE IF NOT EXISTS monitor_notifications (
    id UUID PRIMARY KEY,
    monitor_id UUID NOT NULL,
    notification_channel_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE,
    FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE,
    UNIQUE(monitor_id, notification_channel_id)
)`,
			`CREATE TABLE IF NOT EXISTS monitor_maintenances (
    id UUID PRIMARY KEY,
    monitor_id UUID NOT NULL,
    maintenance_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE,
    FOREIGN KEY (maintenance_id) REFERENCES maintenances(id) ON DELETE CASCADE,
    UNIQUE(monitor_id, maintenance_id)
)`,
			`CREATE TABLE IF NOT EXISTS monitor_status_pages (
    id UUID PRIMARY KEY,
    monitor_id UUID NOT NULL,
    status_page_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE,
    FOREIGN KEY (status_page_id) REFERENCES status_pages(id) ON DELETE CASCADE,
    UNIQUE(monitor_id, status_page_id)
)`,
			`CREATE INDEX IF NOT EXISTS idx_monitors_proxy_id ON monitors(proxy_id)`,
			`CREATE INDEX IF NOT EXISTS idx_status_pages_slug ON status_pages(slug)`,
			`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
			`CREATE INDEX IF NOT EXISTS idx_heartbeats_monitor_time ON heartbeats(monitor_id, check_time)`,
			`CREATE INDEX IF NOT EXISTS idx_stats_monitor_timestamp ON stats(monitor_id, stat_ts)`,
			`CREATE INDEX IF NOT EXISTS idx_monitors_active_status ON monitors(active, status)`,
			`CREATE INDEX IF NOT EXISTS idx_heartbeats_monitor_important ON heartbeats(monitor_id, important)`,
			`CREATE INDEX IF NOT EXISTS idx_heartbeats_monitor_time_important ON heartbeats(monitor_id, check_time, important)`,
			`CREATE INDEX IF NOT EXISTS idx_heartbeats_status ON heartbeats(status)`,
			`CREATE INDEX IF NOT EXISTS idx_heartbeats_important ON heartbeats(important)`,
			`CREATE INDEX IF NOT EXISTS idx_maintenances_active ON maintenances(active)`,
			`CREATE INDEX IF NOT EXISTS idx_status_pages_published ON status_pages(published)`,
			`CREATE INDEX IF NOT EXISTS idx_notification_channels_type ON notification_channels(type)`,
			`CREATE INDEX IF NOT EXISTS idx_notification_channels_active ON notification_channels(active)`,
			`CREATE INDEX IF NOT EXISTS idx_users_active ON users(active)`,
			`CREATE INDEX IF NOT EXISTS idx_proxies_host_port ON proxies(host, port)`,
		}
		for _, stmt := range stmts {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				return err
			}
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		for _, t := range []string{
			"monitor_status_pages", "monitor_maintenances", "monitor_notifications",
			"stats", "heartbeats", "settings", "maintenances",
			"notification_channels", "status_pages", "monitors", "proxies", "users",
		} {
			if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", t)); err != nil {
				return err
			}
		}
		return nil
	})
}
