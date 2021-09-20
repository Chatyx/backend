package config

import (
	"os"
	"sync"

	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const envFilePath = ".env"

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

var (
	cfg  *Config
	once sync.Once
)

func GetConfig(path string) *Config {
	once.Do(func() {
		cfg = &Config{}
		logger := logging.GetLogger()

		if _, err := os.Stat(envFilePath); err == nil {
			if err = godotenv.Load(envFilePath); err != nil {
				logger.WithError(err).Fatal("Failed to loading env variable from %s file", envFilePath)
			}
		}

		if err := cleanenv.ReadConfig(path, cfg); err != nil {
			logger.WithError(err).Fatal("Failed to reading config file %s", path)
		}
	})

	return cfg
}
