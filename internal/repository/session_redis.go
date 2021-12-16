package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/go-redis/redis/v8"
)

type sessionRedisRepository struct {
	redisClient *redis.Client
}

func NewSessionRedisRepository(redisClient *redis.Client) SessionRepository {
	return &sessionRedisRepository{
		redisClient: redisClient,
	}
}

func (r *sessionRedisRepository) Get(ctx context.Context, refreshToken string) (domain.Session, error) {
	key := "session:" + refreshToken

	payload, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return domain.Session{}, fmt.Errorf("%w", domain.ErrSessionNotFound)
		}

		return domain.Session{}, fmt.Errorf("an error occurred while getting refresh session: %v", err)
	}

	var session domain.Session
	if err = json.Unmarshal([]byte(payload), &session); err != nil {
		return domain.Session{}, fmt.Errorf("an error occurred while unmarshalling refresh session payload: %v", err)
	}

	return session, nil
}

func (r *sessionRedisRepository) Set(ctx context.Context, session domain.Session, ttl time.Duration) error {
	sessionKey := "session:" + session.RefreshToken
	userSessionsKey := fmt.Sprintf("user:%s:sessions", session.UserID)

	payload, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("an error occurred while marshaling refresh session: %v", err)
	}

	if err = r.redisClient.Set(ctx, sessionKey, payload, ttl).Err(); err != nil {
		return fmt.Errorf("an error occurred while setting the session: %v", err)
	}

	if err = r.redisClient.RPush(ctx, userSessionsKey, sessionKey).Err(); err != nil {
		return fmt.Errorf("an error occurred while pushing session to the list of user sessions: %v", err)
	}

	return nil
}

func (r *sessionRedisRepository) Delete(ctx context.Context, refreshToken, userID string) error {
	sessionKey := "session:" + refreshToken
	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)

	val, err := r.redisClient.Del(ctx, sessionKey).Result()
	if err != nil {
		return fmt.Errorf("an error occurred while deleting session by key: %v", err)
	}

	if val == 0 {
		return fmt.Errorf("%w", domain.ErrSessionNotFound)
	}

	if err = r.redisClient.LRem(ctx, userSessionsKey, 0, sessionKey).Err(); err != nil {
		return fmt.Errorf("an error occurred while deleting session key from list of user sessions: %v", err)
	}

	return nil
}

func (r *sessionRedisRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)

	keys, err := r.redisClient.LRange(ctx, userSessionsKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("an error occurred while range user session keys: %v", err)
	}

	keys = append(keys, userSessionsKey)

	if err = r.redisClient.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("an error occurred while deleting session keys and key which aggregated them: %v", err)
	}

	return nil
}
