package cmdengine

import (
	"fmt"
	"time"

	"peekaping/internal/config"
	internalengine "peekaping/internal/engine"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config defines the configuration schema for the Engine service.
type Config struct {
	// Database configuration
	DBHost string `env:"DB_HOST"`
	DBPort string `env:"DB_PORT"`
	DBName string `env:"DB_NAME" validate:"required,min=1" default:"monitoring.db"`
	DBUser string `env:"DB_USER"`
	DBPass string `env:"DB_PASS"`
	DBType string `env:"DB_TYPE" validate:"required,db_type" default:"sqlite"`

	// Common settings
	Mode     string `env:"MODE" validate:"required,oneof=dev prod test" default:"dev"`
	LogLevel string `env:"LOG_LEVEL" validate:"omitempty,log_level" default:"info"`
	Timezone string `env:"TZ" validate:"required" default:"UTC"`

	// Redis (optional, only used when ENGINE_USE_REDIS=true)
	RedisHost     string `env:"REDIS_HOST" default:"redis"`
	RedisPort     string `env:"REDIS_PORT" default:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD" default:""`
	RedisDB       int    `env:"REDIS_DB" validate:"min=0,max=15" default:"0"`

	// Engine-specific settings
	EngineWorkers           int           `env:"ENGINE_WORKERS" validate:"min=1" default:"10"`
	EngineSchedulerInterval time.Duration `env:"ENGINE_SCHEDULER_INTERVAL" default:"10s"`
	EngineCheckTimeout      time.Duration `env:"ENGINE_CHECK_TIMEOUT" default:"30s"`
	EngineQueueBuffer       int           `env:"ENGINE_QUEUE_BUFFER" validate:"min=10" default:"1000"`
	EngineUseRedis          bool          `env:"ENGINE_USE_REDIS" default:"false"`

	ServiceName string `env:"SERVICE_NAME" validate:"required,min=1" default:"monitoring:engine"`
}

// LoadAndValidate loads and validates the Engine service configuration.
// v is a pre-configured Viper instance; pass viper.New() when no extra flag bindings are needed.
func LoadAndValidate(path string, v *viper.Viper) (*Config, error) {
	cfg, err := config.LoadWithViper[Config](v, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := Validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate validates the Engine service configuration.
func Validate(cfg *Config) error {
	v := validator.New()
	config.RegisterCustomValidators(v)

	if err := v.Struct(cfg); err != nil {
		return err
	}

	dbConfig := &config.DBConfig{
		DBHost: cfg.DBHost,
		DBPort: cfg.DBPort,
		DBName: cfg.DBName,
		DBUser: cfg.DBUser,
		DBPass: cfg.DBPass,
		DBType: cfg.DBType,
	}
	if err := config.ValidateDatabaseCustomRules(dbConfig); err != nil {
		return fmt.Errorf("database validation failed: %w", err)
	}

	if cfg.EngineUseRedis && cfg.RedisHost == "" {
		return fmt.Errorf("REDIS_HOST is required when ENGINE_USE_REDIS=true")
	}

	return nil
}

// ToInternalConfig converts the engine config to the shared internal config.
func (c *Config) ToInternalConfig() *config.Config {
	return &config.Config{
		DBHost:        c.DBHost,
		DBPort:        c.DBPort,
		DBName:        c.DBName,
		DBUser:        c.DBUser,
		DBPass:        c.DBPass,
		DBType:        c.DBType,
		Mode:          c.Mode,
		LogLevel:      c.LogLevel,
		Timezone:      c.Timezone,
		RedisHost:     c.RedisHost,
		RedisPort:     c.RedisPort,
		RedisPassword: c.RedisPassword,
		RedisDB:       c.RedisDB,
		ServiceName:   c.ServiceName,
	}
}

// ToEngineConfig converts engine-specific settings to the internal engine.Config.
func (c *Config) ToEngineConfig() internalengine.Config {
	return internalengine.Config{
		Workers:           c.EngineWorkers,
		SchedulerInterval: c.EngineSchedulerInterval,
		CheckTimeout:      c.EngineCheckTimeout,
		QueueBuffer:       c.EngineQueueBuffer,
		UseRedis:          c.EngineUseRedis,
	}
}
