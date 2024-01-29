package postgres

import (
	"context"
	"fmt"
	"math"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type MessageRepository struct {
	pool   *pgxpool.Pool
	getter dbClientGetter
}

func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{
		pool:   pool,
		getter: dbClientGetter{pool: pool},
	}
}

func (r *MessageRepository) List(ctx context.Context, obj dto.MessageList) ([]entity.Message, error) {
	b := builder.Select("id", "sender_id", "chat_id", "chat_type",
		"content", "content_type", "is_service", "sent_at", "delivered_at").
		Where(sq.Eq{"chat_id": obj.ChatID.ID, "chat_type": obj.ChatID.Type})

	if obj.Direction == dto.DescDirection {
		if obj.IDAfter == 0 {
			obj.IDAfter = math.MaxInt64
		}
		b = b.Where(sq.Lt{"id": obj.IDAfter}).OrderBy("created_at DESC")
	} else {
		b = b.Where(sq.Gt{"id": obj.IDAfter}).OrderBy("created_at ASC")
	}

	query, args, err := b.Limit(uint64(obj.Limit)).ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select messages query: %v", err)
	}

	rows, err := r.getter.Get(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec query to select messages: %v", err)
	}
	defer rows.Close()

	var messages []entity.Message

	for rows.Next() {
		var message entity.Message

		err = rows.Scan(
			&message.ID, &message.ChatID.ID, &message.ChatID.Type,
			&message.SenderID, &message.Content, &message.ContentType,
			&message.IsService, &message.SentAt, &message.DeliveredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message row: %v", err)
		}

		messages = append(messages, message)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("reading message rows: %v", err)
	}
	return messages, nil
}

func (r *MessageRepository) Create(ctx context.Context, message *entity.Message) error {
	query, args, err := builder.
		Insert("messages").
		Columns("sender_id", "chat_id", "chat_type",
			"content", "content_type", "is_service", "sent_at").
		Values(message.SenderID, message.ChatID.ID, message.ChatID.Type,
			message.Content, message.ContentType, message.IsService, message.SentAt).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert message query: %v", err)
	}

	err = r.getter.Get(ctx).QueryRow(ctx, query, args...).Scan(&message.ID)
	if err != nil {
		return fmt.Errorf("exec query to insert message: %v", err)
	}
	return nil
}
