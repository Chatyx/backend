package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const envFilePath = ".env"

type ListenConfig struct {
	Type     string `yaml:"type"      env-default:"port"`
	BindIP   string `yaml:"bind_ip"   env-default:"127.0.0.1"`
	BindPort int    `yaml:"bind_port" env-default:"8000"`
}

type ListenServers struct {
	API  ListenConfig `yaml:"api"`
	Chat ListenConfig `yaml:"chat"`
}

type AuthConfig struct {
	SignKey         string        `yaml:"sign_key"          env:"SCHT_AUTH_SIGN_KEY" env-required:"true"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl"  env-default:"15"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-default:"43200"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"     env:"SCHT_PG_HOST"     env-required:"true"`
	Port     int    `yaml:"port"     env:"SCHT_PG_PORT"     env-required:"true"`
	Database string `yaml:"database" env:"SCHT_PG_DATABASE" env-required:"true"`
	Username string `yaml:"username" env:"SCHT_PG_USERNAME" env-required:"true"`
	Password string `yaml:"password" env:"SCHT_PG_PASSWORD"`
}

type RedisConfig struct {
	Host     string `yaml:"host"     env:"SCHT_REDIS_HOST"     env-required:"true"`
	Port     int    `yaml:"port"     env:"SCHT_REDIS_PORT"     env-required:"true"`
	Username string `yaml:"username" env:"SCHT_REDIS_USERNAME" env-required:"true"`
	Password string `yaml:"password" env:"SCHT_REDIS_PASSWORD"`
}

type Logging struct {
	Level      string `yaml:"level"       env-default:"debug"`
	FilePath   string `yaml:"filepath"`
	Rotate     bool   `yaml:"rotate"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
}

type Cors struct {
	AllowedOrigins []string `yaml:"allowed_origins" env-default:"*"`
	MaxAge         int      `yaml:"max_age"`
}

type Config struct {
	IsDebug  bool           `yaml:"is_debug" env-default:"false"`
	Domain   string         `yaml:"domain"   env-required:"true"`
	Listen   ListenServers  `yaml:"listen"`
	Auth     AuthConfig     `yaml:"auth"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	Logging  Logging        `yaml:"logging"`
	Cors     Cors           `yaml:"cors"`
}

var (
	cfg  *Config
	once sync.Once
)

func GetConfig(path string) *Config {
	once.Do(func() {
		cfg = &Config{}

		if _, err := os.Stat(envFilePath); err == nil {
			if err = godotenv.Load(envFilePath); err != nil {
				panic(fmt.Sprintf("Failed to loading env variable from %s file: %v", envFilePath, err))
			}
		}

		if err := cleanenv.ReadConfig(path, cfg); err != nil {
			panic(fmt.Sprintf("Failed to reading config file %s: %v", path, err))
		}

		cfg.Auth.AccessTokenTTL *= time.Minute
		cfg.Auth.RefreshTokenTTL *= time.Minute
	})

	return cfg
}

func (c *Config) DBConnectionURL() string {
	pgCfg := c.Postgres

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pgCfg.Username, pgCfg.Password, pgCfg.Host, pgCfg.Port, pgCfg.Database,
	)
}
