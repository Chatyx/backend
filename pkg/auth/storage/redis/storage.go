package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Chatyx/backend/pkg/auth"
	"github.com/Chatyx/backend/pkg/auth/model"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host        string
	Port        string
	Username    string
	Password    string
	DB          int
	ConnTimeout time.Duration
}

type Storage struct {
	cli *redis.Client
}

func NewStorage(conf Config) (*Storage, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Username: conf.Username,
		Password: conf.Password,
		DB:       conf.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), conf.ConnTimeout)
	defer cancel()

	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("establish connection with redis: %v", err)
	}

	return &Storage{cli: cli}, nil
}

func (s *Storage) Set(ctx context.Context, sess model.Session) error {
	sessKey := "session:" + sess.RefreshToken

	cmds, err := s.cli.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, sessKey, Session{
			UserID:      sess.UserID,
			Fingerprint: sess.Fingerprint,
			IP:          sess.IP.String(),
			ExpiresAt:   sess.ExpiresAt.Unix(),
			CreatedAt:   sess.CreatedAt.Unix(),
		})
		pipe.Expire(ctx, sessKey, time.Until(sess.ExpiresAt))

		return nil
	})
	if err != nil {
		return fmt.Errorf("before exec pipeline: %v", err)
	}

	var errs []error
	for _, cmd := range cmds {
		if cmdErr := cmd.Err(); cmdErr != nil {
			errs = append(errs, cmdErr)
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("after exec pipeline: %v", errors.Join(errs...))
	}

	return nil
}

func (s *Storage) GetWithDelete(ctx context.Context, refreshToken string) (model.Session, error) {
	var (
		sess   model.Session
		mapCmd *redis.MapStringStringCmd
	)

	sessKey := "session:" + refreshToken
	cmds, err := s.cli.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		mapCmd = pipe.HGetAll(ctx, sessKey)
		pipe.Del(ctx, sessKey)

		return nil
	})
	if err != nil {
		return sess, fmt.Errorf("before exec pipeline: %v", err)
	}

	var rawSess Session

	if err = mapCmd.Scan(&rawSess); err != nil {
		if errors.Is(err, redis.Nil) {
			return sess, auth.ErrSessionNotFound
		}
		return sess, fmt.Errorf("scan map cmd: %v", err)
	}

	var errs []error
	for _, cmd := range cmds[1:] {
		if cmdErr := cmd.Err(); cmdErr != nil {
			errs = append(errs, cmdErr)
		}
	}

	if len(errs) != 0 {
		return sess, fmt.Errorf("after exec pipeline: %v", errors.Join(errs...))
	}

	return model.Session{
		UserID:       rawSess.UserID,
		RefreshToken: refreshToken,
		Fingerprint:  rawSess.Fingerprint,
		IP:           net.ParseIP(rawSess.IP),
		ExpiresAt:    time.Unix(rawSess.ExpiresAt, 0),
		CreatedAt:    time.Unix(rawSess.CreatedAt, 0),
	}, nil
}
