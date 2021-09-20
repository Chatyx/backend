package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type ListenConfig struct {
	Type     string `yaml:"type"      env-default:"port"`
	BindIP   string `yaml:"bind_ip"   env-default:"127.0.0.1"`
	BindPort int    `yaml:"bind_port" env-default:"8000"`
}

type AuthConfig struct {
	SignKey         string `yaml:"sign_key"          env:"SCHT_AUTH_SIGN_KEY" env-required:"true"`
	AccessTokenTTL  int    `yaml:"access_token_ttl"  env-default:"900"`
	RefreshTokenTTL int    `yaml:"refresh_token_ttl" env-default:"2592000"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"     env:"SCHT_PG_HOST"     env-required:"true"`
	Port     int    `yaml:"port"     env:"SCHT_PG_PORT"     env-required:"true"`
	Database string `yaml:"database" env:"SCHT_PG_DATABASE" env-required:"true"`
	Username string `yaml:"username" env:"SCHT_PG_USERNAME" env-required:"true"`
	Password string `yaml:"password" env:"SCHT_PG_PASSWORD" env-required:"true"`
}

type Config struct {
	IsDebug  bool           `yaml:"is_debug" env-default:"false"`
	Listen   ListenConfig   `yaml:"listen"`
	Auth     AuthConfig     `yaml:"auth"`
	Postgres PostgresConfig `yaml:"postgres"`
}

func GetConfig(path string) (*Config, error) {
	cfg := &Config{}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to get config due %v", err)
	}

	return cfg, nil
}
