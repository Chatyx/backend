package redis

import (
	"context"

	"github.com/Chatyx/backend/internal/entity"

	"github.com/redis/go-redis/v9"
)

type PublishSubscriber struct {
	cli *redis.Client
}

func NewPublishSubscriber(cli *redis.Client) *PublishSubscriber {
	return &PublishSubscriber{cli: cli}
}

func (ps *PublishSubscriber) Publish(ctx context.Context, message entity.Message) error {
	_ = ps.cli
	_, _ = ctx, message
	return nil
}
