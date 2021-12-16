package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

type messagePostgresRepository struct {
	dbPool PgxPool
}

func NewMessagePostgresRepository(dbPool PgxPool) MessageRepository {
	return &messagePostgresRepository{dbPool: dbPool}
}

func (r *messagePostgresRepository) Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error) {
	panic("implement me")
}

func (r *messagePostgresRepository) List(ctx context.Context, chatID string, timestamp time.Time) ([]domain.Message, error) {
	query := `SELECT 
		id, action_id, text, 
		sender_id, chat_id, created_at 
	FROM messages 
	WHERE chat_id = $1 AND created_at > $2 
	ORDER BY created_at`

	rows, err := r.dbPool.Query(ctx, query, chatID, timestamp)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while querying list of messages from database: %v", err)
	}
	defer rows.Close()

	messages := make([]domain.Message, 0)

	for rows.Next() {
		var message domain.Message

		if err = rows.Scan(
			&message.ID, &message.ActionID, &message.Text,
			&message.SenderID, &message.ChatID, &message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("an error occurred while scanning message: %v", err)
		}

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occurred while reading messages: %v", err)
	}

	return messages, nil
}
