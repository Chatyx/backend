package repository

import (
	"context"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/utils"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type messagePostgresRepository struct {
	dbPool PgxPool
	logger logging.Logger
}

func NewMessagePostgresRepository(dbPool PgxPool) MessageRepository {
	return &messagePostgresRepository{
		dbPool: dbPool,
		logger: logging.GetLogger(),
	}
}

func (r *messagePostgresRepository) Store(ctx context.Context, message domain.Message) error {
	panic("implement me")
}

func (r *messagePostgresRepository) List(ctx context.Context, chatID string, timestamp time.Time) ([]domain.Message, error) {
	if !utils.IsValidUUID(chatID) {
		r.logger.Debugf("chat is not found with id = %s", chatID)
		return nil, domain.ErrChatNotFound
	}

	query := `SELECT 
		action, text, sender_id, chat_id, created_at 
	FROM messages 
	WHERE chat_id = $1 AND created_at >= $2 
	ORDER BY created_at`

	rows, err := r.dbPool.Query(ctx, query, chatID, timestamp)
	if err != nil {
		r.logger.WithError(err).Error("Unable to list chat's messages from database")
		return nil, err
	}
	defer rows.Close()

	messages := make([]domain.Message, 0)

	for rows.Next() {
		var message domain.Message

		if err = rows.Scan(
			&message.Action, &message.Text,
			&message.SenderID, &message.ChatID, &message.CreatedAt,
		); err != nil {
			r.logger.WithError(err).Error("Unable to scan message")
			return nil, err
		}

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while reading messages")
		return nil, err
	}

	return messages, nil
}
