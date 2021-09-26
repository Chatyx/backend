package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Mort4lis/scht-backend/pkg/logging"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/go-redis/redis/v8"
)

type sessionRedisRepository struct {
	redisClient *redis.Client
	logger      logging.Logger
}

func NewSessionRedisRepository(redisClient *redis.Client) SessionRepository {
	return &sessionRedisRepository{
		redisClient: redisClient,
		logger:      logging.GetLogger(),
	}
}

func (r *sessionRedisRepository) Get(ctx context.Context, key string) (domain.Session, error) {
	payload, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.logger.WithError(err).Debugf("refresh session is not found with key %s", key)
			return domain.Session{}, domain.ErrSessionNotFound
		}

		r.logger.WithError(err).Error("Error occurred while getting refresh session by key")

		return domain.Session{}, err
	}

	var session domain.Session
	if err = json.Unmarshal([]byte(payload), &session); err != nil {
		r.logger.WithError(err).Error("Error occurred while unmarshalling refresh session payload")
		return domain.Session{}, err
	}

	return session, nil
}

func (r *sessionRedisRepository) Set(ctx context.Context, key string, session domain.Session) error {
	payload, err := json.Marshal(session)
	if err != nil {
		r.logger.WithError(err).Error("Error occurred while marshaling refresh session")
		return err
	}

	ttl := time.Until(session.ExpiresAt)

	if err = r.redisClient.Set(ctx, key, payload, ttl).Err(); err != nil {
		r.logger.WithError(err).Error("Error occurred while setting session to redis")
		return err
	}

	return nil
}

func (r *sessionRedisRepository) Delete(ctx context.Context, key string) error {
	if err := r.redisClient.Del(ctx, key).Err(); err != nil {
		if err == redis.Nil {
			r.logger.WithError(err).Debugf("refresh session is not found with key %s", key)
			return nil
		}

		r.logger.WithError(err).Error("Error occurred while deleting refresh session")

		return err
	}

	return nil
}
