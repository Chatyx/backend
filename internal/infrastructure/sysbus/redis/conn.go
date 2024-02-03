package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Chatyx/backend/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisConn(conf config.Redis) (*redis.Client, error) {
	dbNum, err := strconv.Atoi(conf.Database)
	if err != nil {
		return nil, fmt.Errorf("parse db num: %v", err)
	}

	cli := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Username: conf.User,
		Password: conf.Password,
		DB:       dbNum,
	})

	ctx, cancel := context.WithTimeout(context.Background(), conf.Timeout)
	defer cancel()

	if err = cli.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("establish connection with redis: %v", err)
	}

	return cli, nil
}
