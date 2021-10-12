package repository

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

		r.logger.WithError(err).Error("An error occurred while getting refresh session by key")

		return domain.Session{}, err
	}

	var session domain.Session
	if err = json.Unmarshal([]byte(payload), &session); err != nil {
		r.logger.WithError(err).Error("An error occurred while unmarshalling refresh session payload")
		return domain.Session{}, err
	}

	return session, nil
}

func (r *sessionRedisRepository) Set(ctx context.Context, key string, session domain.Session, ttl time.Duration) error {
	payload, err := json.Marshal(session)
	if err != nil {
		r.logger.WithError(err).Error("An error occurred while marshaling refresh session")
		return err
	}

	if err = r.redisClient.Set(ctx, key, payload, ttl).Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while setting session")
		return err
	}

	if err = r.redisClient.RPush(ctx, session.UserID, key).Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while pushing session")
		return err
	}

	return nil
}

func (r *sessionRedisRepository) Delete(ctx context.Context, key, userID string) error {
	val, err := r.redisClient.Del(ctx, key).Result()
	if err != nil {
		r.logger.WithError(err).Error("An error occurred while deleting refresh session")
		return err
	}

	if val == 0 {
		r.logger.WithError(err).Debugf("refresh session is not found with key %s", key)
		return domain.ErrSessionNotFound
	}

	if err = r.redisClient.LRem(ctx, userID, 0, key).Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while deleting refresh session")
		return err
	}

	return nil
}

func (r *sessionRedisRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	keys, err := r.redisClient.LRange(ctx, userID, 0, -1).Result()
	if err != nil {
		r.logger.WithError(err).Error("An error occurred while range session keys by user_id")
		return err
	}

	keys = append(keys, userID)

	if err = r.redisClient.Del(ctx, keys...).Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while deleting session keys by user_id")
		return err
	}

	return nil
}
