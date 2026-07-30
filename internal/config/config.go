package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
}

type ServerConfig struct {
	ListenAddr   string   `yaml:"listen_addr"`
	AllowOrigins []string `yaml:"allow_origins"`
}

type DatabaseConfig struct {
	Main string `yaml:"main"`
}

type AuthConfig struct {
	JWTSecret                string `yaml:"jwt_secret"`
	AccessTokenExpireSeconds  int    `yaml:"access_token_expire_seconds"`
	RefreshTokenExpireSeconds int    `yaml:"refresh_token_expire_seconds"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	setDefaults(&cfg)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = ":8080"
	}
	if cfg.Auth.AccessTokenExpireSeconds == 0 {
		cfg.Auth.AccessTokenExpireSeconds = 900
	}
	if cfg.Auth.RefreshTokenExpireSeconds == 0 {
		cfg.Auth.RefreshTokenExpireSeconds = 604800
	}
}

func (c *Config) Validate() error {
	if c.Database.Main == "" {
		return fmt.Errorf("database.main is required")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}
	return nil
}
