package repository

import (
	"bytes"
	"context"
	"fmt"

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

func (r *messagePostgresRepository) List(ctx context.Context, chatID string, dto domain.MessageListDTO) (domain.MessageList, error) {
	query := r.buildListQuery(dto.Direction)

	rows, err := r.dbPool.Query(ctx, query, chatID, dto.OffsetDate, dto.Offset, dto.Limit)
	if err != nil {
		return domain.MessageList{}, fmt.Errorf("an error occurred while querying list of messages from database: %v", err)
	}
	defer rows.Close()

	messages := make([]domain.Message, 0)

	for rows.Next() {
		var message domain.Message

		if err = rows.Scan(
			&message.ID, &message.ActionID, &message.Text,
			&message.SenderID, &message.ChatID, &message.CreatedAt,
		); err != nil {
			return domain.MessageList{}, fmt.Errorf("an error occurred while scanning message: %v", err)
		}

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return domain.MessageList{}, fmt.Errorf("an error occurred while reading messages: %v", err)
	}

	total, err := r.getTotalMessages(ctx, chatID, dto)
	if err != nil {
		return domain.MessageList{}, err
	}

	return domain.MessageList{
		Total:    total,
		Messages: messages,
	}, nil
}

func (r *messagePostgresRepository) getTotalMessages(ctx context.Context, chatID string, dto domain.MessageListDTO) (int, error) {
	total := 0
	builder := bytes.NewBufferString(`SELECT COUNT(*) as total FROM messages WHERE chat_id = $1 AND `)

	if dto.Direction == domain.NewerMessages {
		builder.WriteString("created_at >= $2")
	} else {
		builder.WriteString("created_at <= $2")
	}

	if err := r.dbPool.QueryRow(ctx, builder.String(), chatID, dto.OffsetDate).Scan(&total); err != nil {
		return total, fmt.Errorf("an error occurred while getting total messages: %v", err)
	}

	return total, nil
}

func (r *messagePostgresRepository) buildListQuery(direction string) string {
	builder := bytes.NewBufferString(`SELECT 
		id, action_id, text, 
		sender_id, chat_id, created_at 
	FROM messages 
	WHERE chat_id = $1 AND `)

	if direction == domain.NewerMessages {
		builder.WriteString("created_at >= $2")
	} else {
		builder.WriteString("created_at <= $2")
	}

	builder.WriteString(" ORDER BY created_at ")

	if direction == domain.NewerMessages {
		builder.WriteString("ASC")
	} else {
		builder.WriteString("DESC")
	}

	builder.WriteString(" OFFSET $3 LIMIT $4")

	return builder.String()
}
