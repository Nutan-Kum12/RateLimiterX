package configs

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig          `mapstructure:"server"`
	Database DatabaseConfig        `mapstructure:"database"`
	Redis    RedisConfig           `mapstructure:"redis"`
	JWT      JWTConfig             `mapstructure:"jwt"`
	Tiers    map[string]TierConfig `mapstructure:"tiers"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // "debug" | "release" | "test"
}

// DatabaseConfig holds MySQL connection settings.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// DSN(Data Source Name) returns the MySQL data source name.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig holds JWT signing settings.
type JWTConfig struct {
	AccessSecret  string        `mapstructure:"access_secret"`
	RefreshSecret string        `mapstructure:"refresh_secret"`
	AccessTTL     time.Duration `mapstructure:"access_ttl"`
	RefreshTTL    time.Duration `mapstructure:"refresh_ttl"`
}

// TierConfig holds rate-limiting settings for a user tier.
type TierConfig struct {
	Algorithm string        `mapstructure:"algorithm"` // "fixed_window", "sliding_window", "sliding_log", "token_bucket"
	Limit     int           `mapstructure:"limit"`     // Max requests per window
	Window    time.Duration `mapstructure:"window"`    // Time window duration
	Burst     int           `mapstructure:"burst"`     // Token bucket burst capacity
}

// Load reads configuration from the YAML file and environment variables.
// It looks for config/config.yaml relative to the working directory.
func Load(configPath string) (*Config, error) {
	_ = godotenv.Load()
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Allow environment variables to override config values.
	v.SetEnvPrefix("RATELIMITERX")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set sensible defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.user", "root")
	v.SetDefault("database.name", "ratelimiterx")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.access_ttl", "15m")
	v.SetDefault("jwt.refresh_ttl", "168h")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// validate checks that critical configuration values are present.
func (c *Config) validate() error {
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("jwt.access_secret is required")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("jwt.refresh_secret is required")
	}
	if len(c.Tiers) == 0 {
		return fmt.Errorf("at least one tier must be configured")
	}
	for name, tier := range c.Tiers {
		validAlgorithms := map[string]bool{
			"fixed_window":   true,
			"sliding_window": true,
			"sliding_log":    true,
			"token_bucket":   true,
		}
		if !validAlgorithms[tier.Algorithm] {
			return fmt.Errorf("tier %q has invalid algorithm %q", name, tier.Algorithm)
		}
		if tier.Limit <= 0 {
			return fmt.Errorf("tier %q limit must be positive", name)
		}
		if tier.Window <= 0 {
			return fmt.Errorf("tier %q window must be positive", name)
		}
	}
	return nil
}
