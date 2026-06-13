package db

import "time"

// AllModels returns all GORM models in dependency order for AutoMigrate.
func AllModels() []any {
	return []any{
		&Proxy{},
		&User{},
		&StatusPage{},
		&NotificationChannel{},
		&Maintenance{},
		&Tag{},
		&Setting{},
		&LoginState{},
		&NotificationSentHistory{},
		&APIKey{},
		&Monitor{},
		&HeartBeat{},
		&Stat{},
		&MonitorNotification{},
		&MonitorMaintenance{},
		&MonitorStatusPage{},
		&MonitorTag{},
		&DomainStatusPage{},
		&MonitorTLSInfo{},
	}
}

type User struct {
	ID             string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Email          string    `gorm:"column:email;not null;uniqueIndex;type:varchar(255)"`
	Password       string    `gorm:"column:password;not null;type:varchar(255)"`
	Active         bool      `gorm:"column:active;not null;default:true"`
	TwoFASecret    string    `gorm:"column:twofa_secret;type:varchar(64)"`
	TwoFAStatus    bool      `gorm:"column:twofa_status;not null;default:false"`
	TwoFALastToken string    `gorm:"column:twofa_last_token;type:varchar(6)"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (User) TableName() string { return "users" }

type Proxy struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Protocol  string    `gorm:"column:protocol;not null;type:varchar(10)"`
	Host      string    `gorm:"column:host;not null;type:varchar(255)"`
	Port      int       `gorm:"column:port;not null"`
	Auth      bool      `gorm:"column:auth;not null;default:false"`
	Username  string    `gorm:"column:username;type:varchar(255)"`
	Password  string    `gorm:"column:password;type:varchar(255)"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (Proxy) TableName() string { return "proxies" }

type Monitor struct {
	ID             string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Type           string    `gorm:"column:type;not null;type:varchar(20)"`
	Name           string    `gorm:"column:name;not null;type:varchar(150)"`
	Interval       int       `gorm:"column:check_interval;not null"` // 'interval' reserved in MySQL
	Timeout        int       `gorm:"column:timeout;not null"`
	MaxRetries     int       `gorm:"column:max_retries;not null"`
	RetryInterval  int       `gorm:"column:retry_interval;not null"`
	ResendInterval int       `gorm:"column:resend_interval;not null"`
	Active         bool      `gorm:"column:active;not null;default:true;index:idx_monitors_active_status"`
	Status         int       `gorm:"column:status;not null;default:0;index:idx_monitors_active_status"`
	Config         string    `gorm:"column:config;type:json"`
	ProxyID        *string   `gorm:"column:proxy_id;type:varchar(36);index:idx_monitors_proxy_id"`
	PushToken      string    `gorm:"column:push_token;type:varchar(255)"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
	Proxy          *Proxy    `gorm:"foreignKey:ProxyID;references:ID;constraint:OnDelete:SET NULL"`
}

func (Monitor) TableName() string { return "monitors" }

type StatusPage struct {
	ID                  string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Slug                string    `gorm:"column:slug;not null;uniqueIndex;type:varchar(255)"`
	Title               string    `gorm:"column:title;not null;type:varchar(255)"`
	Description         string    `gorm:"column:description;type:text"`
	Icon                string    `gorm:"column:icon;type:varchar(255)"`
	Theme               string    `gorm:"column:theme;not null;default:light;type:varchar(30)"`
	Published           bool      `gorm:"column:published;not null;default:false;index:idx_status_pages_published"`
	FooterText          string    `gorm:"column:footer_text;type:text"`
	AutoRefreshInterval int       `gorm:"column:auto_refresh_interval;not null;default:300"`
	CreatedAt           time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (StatusPage) TableName() string { return "status_pages" }

type NotificationChannel struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Name      string    `gorm:"column:name;not null;type:varchar(255)"`
	Type      string    `gorm:"column:type;not null;type:varchar(50);index:idx_notification_channels_type"`
	Active    bool      `gorm:"column:active;not null;default:true;index:idx_notification_channels_active"`
	IsDefault bool      `gorm:"column:is_default;not null;default:false"`
	Config    *string   `gorm:"column:config;type:json"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (NotificationChannel) TableName() string { return "notification_channels" }

type Maintenance struct {
	ID            string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Title         string    `gorm:"column:title;not null;type:varchar(255)"`
	Description   string    `gorm:"column:description;type:text"`
	Active        bool      `gorm:"column:active;not null;default:true;index:idx_maintenances_active"`
	Strategy      string    `gorm:"column:strategy;not null;type:varchar(50)"`
	StartDateTime *string   `gorm:"column:start_date_time;type:varchar(50)"`
	EndDateTime   *string   `gorm:"column:end_date_time;type:varchar(50)"`
	StartTime     *string   `gorm:"column:start_time;type:varchar(50)"`
	EndTime       *string   `gorm:"column:end_time;type:varchar(50)"`
	Weekdays      string    `gorm:"column:weekdays;type:text"`
	DaysOfMonth   string    `gorm:"column:days_of_month;type:text"`
	IntervalDay   *int      `gorm:"column:interval_day"`
	Cron          *string   `gorm:"column:cron;type:varchar(255)"`
	Timezone      *string   `gorm:"column:timezone;type:varchar(100)"`
	Duration      *int      `gorm:"column:duration"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (Maintenance) TableName() string { return "maintenances" }

// 'key' is reserved in MySQL — renamed to setting_key
type Setting struct {
	SettingKey string    `gorm:"column:setting_key;primaryKey;type:varchar(255)"`
	Value      string    `gorm:"column:value;not null;type:text"`
	Type       string    `gorm:"column:type;not null;type:varchar(50)"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (Setting) TableName() string { return "settings" }

// 'time' is reserved in MySQL — renamed to check_time
type HeartBeat struct {
	ID        string     `gorm:"column:id;primaryKey;type:varchar(36)"`
	MonitorID string     `gorm:"column:monitor_id;not null;index:idx_heartbeats_monitor_time,composite:monitor_time;index:idx_heartbeats_monitor_important,composite:monitor_important;index:idx_heartbeats_monitor_time_important,composite:monitor_time_important"`
	Status    int        `gorm:"column:status;not null;index:idx_heartbeats_status"`
	Msg       string     `gorm:"column:msg;type:text"`
	Ping      int        `gorm:"column:ping"`
	Duration  int        `gorm:"column:duration"`
	DownCount int        `gorm:"column:down_count"`
	Retries   int        `gorm:"column:retries"`
	Important bool       `gorm:"column:important;not null;default:false;index:idx_heartbeats_important;index:idx_heartbeats_monitor_important,composite:monitor_important;index:idx_heartbeats_monitor_time_important,composite:monitor_time_important"`
	CheckTime time.Time  `gorm:"column:check_time;not null;default:CURRENT_TIMESTAMP;index:idx_heartbeats_monitor_time,composite:monitor_time;index:idx_heartbeats_monitor_time_important,composite:monitor_time_important"`
	EndTime   *time.Time `gorm:"column:end_time"`
	Notified  bool       `gorm:"column:notified;not null;default:false"`
	Monitor   *Monitor   `gorm:"foreignKey:MonitorID;references:ID;constraint:OnDelete:CASCADE"`
}

func (HeartBeat) TableName() string { return "heartbeats" }

// 'timestamp' is reserved in MySQL — renamed to stat_ts
type Stat struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	MonitorID   string    `gorm:"column:monitor_id;not null;uniqueIndex:idx_stats_unique;index:idx_stats_monitor_timestamp"`
	StatTS      time.Time `gorm:"column:stat_ts;not null;uniqueIndex:idx_stats_unique;index:idx_stats_monitor_timestamp"`
	Ping        float64   `gorm:"column:ping;not null;default:0"`
	PingMin     float64   `gorm:"column:ping_min;not null;default:0"`
	PingMax     float64   `gorm:"column:ping_max;not null;default:0"`
	Up          int       `gorm:"column:up;not null;default:0"`
	Down        int       `gorm:"column:down;not null;default:0"`
	Maintenance int       `gorm:"column:maintenance;not null;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
	Monitor     *Monitor  `gorm:"foreignKey:MonitorID;references:ID;constraint:OnDelete:CASCADE"`
}

func (Stat) TableName() string { return "stats" }

type MonitorNotification struct {
	ID                    string               `gorm:"column:id;primaryKey;type:varchar(36)"`
	MonitorID             string               `gorm:"column:monitor_id;not null;uniqueIndex:idx_monitor_notification_unique"`
	NotificationChannelID string               `gorm:"column:notification_channel_id;not null;uniqueIndex:idx_monitor_notification_unique"`
	CreatedAt             time.Time            `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt             time.Time            `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
	Monitor               *Monitor             `gorm:"foreignKey:MonitorID;references:ID;constraint:OnDelete:CASCADE"`
	NotificationChannel   *NotificationChannel `gorm:"foreignKey:NotificationChannelID;references:ID;constraint:OnDelete:CASCADE"`
}

func (MonitorNotification) TableName() string { return "monitor_notifications" }

type MonitorMaintenance struct {
	ID            string       `gorm:"column:id;primaryKey;type:varchar(36)"`
	MonitorID     string       `gorm:"column:monitor_id;not null;uniqueIndex:idx_monitor_maintenance_unique"`
	MaintenanceID string       `gorm:"column:maintenance_id;not null;uniqueIndex:idx_monitor_maintenance_unique"`
	CreatedAt     time.Time    `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time    `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
	Monitor       *Monitor     `gorm:"foreignKey:MonitorID;references:ID;constraint:OnDelete:CASCADE"`
	Maintenance   *Maintenance `gorm:"foreignKey:MaintenanceID;references:ID;constraint:OnDelete:CASCADE"`
}

func (MonitorMaintenance) TableName() string { return "monitor_maintenances" }

type MonitorStatusPage struct {
	ID           string      `gorm:"column:id;primaryKey;type:varchar(36)"`
	MonitorID    string      `gorm:"column:monitor_id;not null;uniqueIndex:idx_monitor_status_page_unique"`
	StatusPageID string      `gorm:"column:status_page_id;not null;uniqueIndex:idx_monitor_status_page_unique"`
	CreatedAt    time.Time   `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time   `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
	Monitor      *Monitor    `gorm:"foreignKey:MonitorID;references:ID;constraint:OnDelete:CASCADE"`
	StatusPage   *StatusPage `gorm:"foreignKey:StatusPageID;references:ID;constraint:OnDelete:CASCADE"`
}

func (MonitorStatusPage) TableName() string { return "monitor_status_pages" }

type Tag struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Name        string    `gorm:"column:name;not null;uniqueIndex:idx_tags_name;type:varchar(100)"`
	Color       string    `gorm:"column:color;not null;default:#3B82F6;type:varchar(7)"`
	Description *string   `gorm:"column:description;type:text"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (Tag) TableName() string { return "tags" }

type MonitorTag struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	MonitorID string    `gorm:"column:monitor_id;not null;uniqueIndex:idx_monitor_tag_unique;index:idx_monitor_tags_monitor_id;index:idx_monitor_tags_tag_monitor"`
	TagID     string    `gorm:"column:tag_id;not null;uniqueIndex:idx_monitor_tag_unique;index:idx_monitor_tags_tag_id;index:idx_monitor_tags_tag_monitor"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
	Monitor   *Monitor  `gorm:"foreignKey:MonitorID;references:ID;constraint:OnDelete:CASCADE"`
	Tag       *Tag      `gorm:"foreignKey:TagID;references:ID;constraint:OnDelete:CASCADE"`
}

func (MonitorTag) TableName() string { return "monitor_tags" }

type LoginState struct {
	LookupKey   string     `gorm:"column:lookup_key;primaryKey;type:varchar(255)"`
	FailCount   int        `gorm:"column:fail_count;not null"`
	FirstFailAt time.Time  `gorm:"column:first_fail_at;not null"`
	LockedUntil *time.Time `gorm:"column:locked_until;index:login_state_locked_until_idx"`
}

func (LoginState) TableName() string { return "login_state" }

type DomainStatusPage struct {
	ID           string      `gorm:"column:id;primaryKey;type:varchar(36)"`
	StatusPageID string      `gorm:"column:status_page_id;not null;uniqueIndex:idx_domain_status_page_unique"`
	Domain       string      `gorm:"column:domain;not null;type:text;uniqueIndex:idx_domain_status_page_unique"`
	CreatedAt    time.Time   `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time   `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
	StatusPage   *StatusPage `gorm:"foreignKey:StatusPageID;references:ID;constraint:OnDelete:CASCADE"`
}

func (DomainStatusPage) TableName() string { return "domain_status_page" }

type NotificationSentHistory struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Type      string    `gorm:"column:type;not null;type:varchar(50);index:idx_notification_sent_history_type_monitor"`
	MonitorID string    `gorm:"column:monitor_id;not null;type:varchar(255);index:idx_notification_sent_history_type_monitor;uniqueIndex:idx_notification_sent_history_unique"`
	Days      int       `gorm:"column:days;not null;uniqueIndex:idx_notification_sent_history_unique"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;index:idx_notification_sent_history_created_at"`
}

func (NotificationSentHistory) TableName() string { return "notification_sent_history" }

type MonitorTLSInfo struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	MonitorID string    `gorm:"column:monitor_id;not null;uniqueIndex:idx_monitor_tls_info_monitor_id;type:varchar(255)"`
	InfoJSON  string    `gorm:"column:info_json;not null;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;index:idx_monitor_tls_info_updated_at"`
}

func (MonitorTLSInfo) TableName() string { return "monitor_tls_info" }

type APIKey struct {
	ID            string     `gorm:"column:id;primaryKey;type:varchar(36)"`
	Name          string     `gorm:"column:name;not null;type:varchar(255)"`
	KeyHash       *string    `gorm:"column:key_hash;type:varchar(255);index:idx_api_keys_key_hash"`
	DisplayKey    string     `gorm:"column:display_key;default:pk_****;type:varchar(20)"`
	LastUsed      *time.Time `gorm:"column:last_used"`
	ExpiresAt     *time.Time `gorm:"column:expires_at;index:idx_api_keys_expires_at"`
	UsageCount    int64      `gorm:"column:usage_count;not null;default:0"`
	MaxUsageCount *int64     `gorm:"column:max_usage_count"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (APIKey) TableName() string { return "api_keys" }
