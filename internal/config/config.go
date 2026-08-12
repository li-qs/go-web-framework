package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

const (
	defaultServerAddr   = ":8080"
	defaultAccessTTL    = 900
	defaultRefreshTTL   = 604800
	defaultCookieSecure = true
)

type Config struct {
	ServerAddr   string   `yaml:"server_addr"`
	AllowOrigins []string `yaml:"allow_origins"`
	LogLevel     string   `yaml:"log_level"`
	MySQLDSN     string   `yaml:"mysql_dsn"`
	JWTSecret    string   `yaml:"jwt_secret"`
	TokenSalt    string   `yaml:"token_salt"`
	AccessTTL    int      `yaml:"access_ttl"`
	RefreshTTL   int      `yaml:"refresh_ttl"`
	CookieSecure *bool    `yaml:"cookie_secure"`
}

func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = defaultServerAddr
	}
	if cfg.AccessTTL == 0 {
		cfg.AccessTTL = defaultAccessTTL
	}
	if cfg.RefreshTTL == 0 {
		cfg.RefreshTTL = defaultRefreshTTL
	}
	if cfg.CookieSecure == nil {
		secure := defaultCookieSecure
		cfg.CookieSecure = &secure
	}

	if cfg.MySQLDSN == "" {
		return nil, fmt.Errorf("mysql_dsn is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("jwt_secret is required")
	}

	return &cfg, nil
}
