package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	core "github.com/Chatyx/backend/pkg/auth"

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

func (s *Storage) Set(ctx context.Context, sess core.Session) error {
	sessKey := "session:" + sess.RefreshToken
	userSessKey := fmt.Sprintf("user:%s:sessions", sess.UserID)

	cmds, err := s.cli.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, sessKey, Session{
			UserID:      sess.UserID,
			Fingerprint: sess.Fingerprint,
			IP:          sess.IP.String(),
			ExpiresAt:   sess.ExpiresAt.Unix(),
			CreatedAt:   sess.CreatedAt.Unix(),
		})
		pipe.SAdd(ctx, userSessKey, sess.RefreshToken)

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

func (s *Storage) GetWithDelete(ctx context.Context, refreshToken string) (core.Session, error) {
	var (
		sess    core.Session
		rawSess Session
		hGetCmd *redis.MapStringStringCmd
	)

	sessKey := "session:" + refreshToken
	cmds, err := s.cli.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		hGetCmd = pipe.HGetAll(ctx, sessKey)
		pipe.Del(ctx, sessKey)

		return nil
	})
	if err != nil {
		return sess, fmt.Errorf("before exec pipeline: %v", err)
	}

	if err = hGetCmd.Scan(&rawSess); err != nil {
		return sess, fmt.Errorf("scan map cmd: %v", err)
	}

	if rawSess.UserID == "" {
		return sess, core.ErrSessionNotFound
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

	userSessKey := fmt.Sprintf("user:%s:sessions", rawSess.UserID)
	if err = s.cli.SRem(ctx, userSessKey, refreshToken).Err(); err != nil {
		return sess, fmt.Errorf("remove element from set: %v", err)
	}

	return core.Session{
		UserID:       rawSess.UserID,
		RefreshToken: refreshToken,
		Fingerprint:  rawSess.Fingerprint,
		IP:           net.ParseIP(rawSess.IP),
		ExpiresAt:    time.Unix(rawSess.ExpiresAt, 0).Local(),
		CreatedAt:    time.Unix(rawSess.CreatedAt, 0).Local(),
	}, nil
}

func (s *Storage) DeleteAllByUserID(ctx context.Context, id string) error {
	_, _ = ctx, id
	panic("implement me")
}

func (s *Storage) Close() error {
	if err := s.cli.Close(); err != nil {
		return fmt.Errorf("redis client close: %v", err)
	}
	return nil
}
