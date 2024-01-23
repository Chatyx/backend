package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/Chatyx/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	foreignKeyViolationCode = "23503"
	uniqueViolationCode     = "23505"
)

func buildConnString(conf config.Postgres) string {
	b := strings.Builder{}
	b.WriteString(
		fmt.Sprintf(
			"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
			conf.User, conf.Password, conf.Host, conf.Port, conf.Database,
		),
	)

	if conf.MaxOpenConns != 0 {
		b.WriteString(fmt.Sprintf(" pool_max_conns=%d", conf.MaxOpenConns))
	}
	if conf.MinOpenConns != 0 {
		b.WriteString(fmt.Sprintf(" pool_min_conns=%d", conf.MinOpenConns))
	}
	if conf.ConnMaxLifetime != 0 {
		b.WriteString(fmt.Sprintf(" pool_max_conn_lifetime=%s", conf.ConnMaxLifetime))
	}
	if conf.ConnMaxIdleTime != 0 {
		b.WriteString(fmt.Sprintf(" pool_max_conn_idle_time=%s", conf.ConnMaxIdleTime))
	}

	return b.String()
}

func NewConnPool(conf config.Postgres) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), conf.Timeout)
	defer cancel()

	poolConf, err := pgxpool.ParseConfig(buildConnString(conf))
	if err != nil {
		return nil, fmt.Errorf("parse connection string: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConf)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %v", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %v", err)
	}
	return pool, nil
}
