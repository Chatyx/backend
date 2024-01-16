package config

import (
	"time"
)

type Log struct {
	Level string `env-default:"debug" yaml:"level"`
}

type Server struct {
	Listen       string        `env:"LISTEN"      env-default:":8080"  yaml:"listen"`
	ReadTimeout  time.Duration `env-default:"15s" yaml:"read_timeout"`
	WriteTimeout time.Duration `env-default:"15s" yaml:"write_timeout"`
}

type Cors struct {
	AllowedOrigins []string      `env-default:"*" yaml:"allowed_origins"`
	MaxAge         time.Duration `yaml:"max_age"`
}

type Auth struct {
	Issuer          string        `env-default:"chatyx" yaml:"issuer"`
	SignKey         string        `env:"SIGN_KEY"       env-required:"true"      yaml:"sign_key"`
	AccessTokenTTL  time.Duration `env-default:"15m"    yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `env-default:"720h"   yaml:"refresh_token_ttl"`
}

type Conn struct {
	Host     string        `env:"HOST"        env-default:"localhost" yaml:"host"`
	Port     string        `env:"PORT"        env-required:"true"     yaml:"port"`
	Database string        `env:"DB"          env-required:"true"     yaml:"database"`
	User     string        `env:"USER"        env-required:"true"     yaml:"user"`
	Password string        `env:"PASSWORD"    yaml:"password"`
	Timeout  time.Duration `env-default:"15s" yaml:"timeout"`
}

type Postgres struct {
	Conn `yaml:"conn"`

	// MaxOpenConns the maximum number of open connections to the database.
	//
	// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
	// MaxIdleConns, then MaxIdleConns will be reduced to match the new
	// MaxOpenConns limit.
	//
	// If n <= 0, then there is no limit on the number of open connections.
	// The default is 0 (unlimited).
	MaxOpenConns int `yaml:"max_open_conns"`

	// MaxIdleConns the maximum number of connections in the idle
	// connection pool.
	//
	// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
	// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
	//
	// If n <= 0, no idle connections are retained.
	//
	// The default max idle connections is currently 2. This may change in
	// a future.
	MaxIdleConns int `yaml:"max_idle_conns"`

	// ConnMaxIdleTime the maximum amount of time a connection may be idle.
	//
	// Expired connections may be closed lazily before reuse.
	//
	// If d <= 0, connections are not closed due to a connection's idle time.
	// The default is 0 (unlimited).
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`

	// ConnMaxLifetime the maximum amount of time a connection may be reused.
	//
	// Expired connections may be closed lazily before reuse.
	//
	// If d <= 0, connections are not closed due to a connection's age.
	// The default is 0 (unlimited).
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type Redis struct {
	Conn `yaml:"conn"`
}

type Config struct {
	Domain   string   `env-default:"localhost" yaml:"domain"`
	Debug    bool     `yaml:"debug"`
	Log      Log      `yaml:"log"`
	API      Server   `yaml:"api"`
	Chat     Server   `yaml:"chat"`
	Cors     Cors     `yaml:"cors"`
	Auth     Auth     `yaml:"auth"`
	Postgres Postgres `env-prefix:"POSTGRES_"  yaml:"postgres"`
	Redis    Redis    `env-prefix:"REDIS_"     yaml:"redis"`
}
