package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
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

func (r *sessionRedisRepository) Get(ctx context.Context, refreshToken string) (domain.Session, error) {
	key := "session:" + refreshToken

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

func (r *sessionRedisRepository) Set(ctx context.Context, session domain.Session, ttl time.Duration) error {
	sessionKey := "session:" + session.RefreshToken
	userSessionsKey := fmt.Sprintf("user:%s:sessions", session.UserID)

	payload, err := json.Marshal(session)
	if err != nil {
		r.logger.WithError(err).Error("An error occurred while marshaling refresh session")
		return err
	}

	if err = r.redisClient.Set(ctx, sessionKey, payload, ttl).Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while setting session")
		return err
	}

	if err = r.redisClient.RPush(ctx, userSessionsKey, sessionKey).Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while pushing session to the list of user's sessions")
		return err
	}

	return nil
}

func (r *sessionRedisRepository) Delete(ctx context.Context, refreshToken, userID string) error {
	sessionKey := "session:" + refreshToken
	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)

	val, err := r.redisClient.Del(ctx, sessionKey).Result()
	if err != nil {
		r.logger.WithError(err).Error("An error occurred while deleting session by key")
		return err
	}

	if val == 0 {
		r.logger.WithError(err).Debugf("refresh session is not found with key %s", sessionKey)
		return domain.ErrSessionNotFound
	}

	if err = r.redisClient.LRem(ctx, userSessionsKey, 0, sessionKey).Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while deleting session key from list of user's sessions")
		return err
	}

	return nil
}

func (r *sessionRedisRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)

	keys, err := r.redisClient.LRange(ctx, userSessionsKey, 0, -1).Result()
	if err != nil {
		r.logger.WithError(err).Error("An error occurred while range user's session keys")
		return err
	}

	keys = append(keys, userSessionsKey)

	if err = r.redisClient.Del(ctx, keys...).Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while deleting session keys and key which aggregated them")
		return err
	}

	return nil
}
