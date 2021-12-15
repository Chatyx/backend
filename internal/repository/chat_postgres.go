package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/jackc/pgx/v4"
)

type chatPostgresRepository struct {
	dbPool PgxPool
}

func NewChatPostgresRepository(dbPool PgxPool) ChatRepository {
	return &chatPostgresRepository{dbPool: dbPool}
}

func (r *chatPostgresRepository) List(ctx context.Context, userID string) ([]domain.Chat, error) {
	query := `SELECT 
		chats.id, chats.name, chats.description, 
		chats.creator_id, chats.created_at, chats.updated_at 
	FROM chats 
	INNER JOIN chat_members 
		ON chats.id = chat_members.chat_id
	WHERE chat_members.user_id = $1`

	rows, err := r.dbPool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while querying list of chats from database: %v", err)
	}
	defer rows.Close()

	chats := make([]domain.Chat, 0)

	for rows.Next() {
		var chat domain.Chat

		if err = rows.Scan(
			&chat.ID, &chat.Name, &chat.Description,
			&chat.CreatorID, &chat.CreatedAt, &chat.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("an error occurred while scanning chat: %v", err)
		}

		chats = append(chats, chat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occurred while reading chats: %v", err)
	}

	return chats, nil
}

func (r *chatPostgresRepository) Create(ctx context.Context, dto domain.CreateChatDTO) (domain.Chat, error) {
	var chat domain.Chat

	err := r.dbPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		query := "INSERT INTO chats (name, description, creator_id) VALUES ($1, $2, $3) RETURNING id, created_at"

		if err := r.dbPool.QueryRow(
			ctx, query,
			dto.Name, dto.Description, dto.CreatorID,
		).Scan(&chat.ID, &chat.CreatedAt); err != nil {
			return fmt.Errorf("an error occurred while creating chat into the database: %v", err)
		}

		query = "INSERT INTO chat_members (user_id, chat_id) VALUES ($1, $2)"

		if _, err := r.dbPool.Exec(ctx, query, dto.CreatorID, chat.ID); err != nil {
			return fmt.Errorf("an error occurred while inserting into chat_members table: %v", err)
		}

		return nil
	})
	if err != nil {
		return domain.Chat{}, err
	}

	chat.Name = dto.Name
	chat.Description = dto.Description
	chat.CreatorID = dto.CreatorID

	return chat, nil
}

func (r *chatPostgresRepository) Get(ctx context.Context, memberKey domain.ChatMemberIdentity) (domain.Chat, error) {
	query := `SELECT 
		chats.id, chats.name, chats.description, 
		chats.creator_id, chats.created_at, chats.updated_at 
	FROM chats 
	INNER JOIN chat_members 
		ON chats.id = chat_members.chat_id
	WHERE chats.id = $1 AND chat_members.user_id = $2`

	return r.getOne(ctx, query, memberKey.ChatID, memberKey.UserID)
}

func (r *chatPostgresRepository) getOne(ctx context.Context, query string, args ...interface{}) (domain.Chat, error) {
	var chat domain.Chat

	if err := r.dbPool.QueryRow(ctx, query, args...).Scan(
		&chat.ID, &chat.Name, &chat.Description,
		&chat.CreatorID, &chat.CreatedAt, &chat.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Chat{}, fmt.Errorf("%w", domain.ErrChatNotFound)
		}

		return domain.Chat{}, fmt.Errorf("an error occurred while getting chat: %v", err)
	}

	return chat, nil
}

func (r *chatPostgresRepository) Update(ctx context.Context, dto domain.UpdateChatDTO) (domain.Chat, error) {
	query := `UPDATE chats 
	SET name = $1, description = $2 
	WHERE id = $3 AND creator_id = $4 
	RETURNING created_at, updated_at`

	var chat domain.Chat

	if err := r.dbPool.QueryRow(
		ctx, query,
		dto.Name, dto.Description,
		dto.ID, dto.CreatorID,
	).Scan(&chat.CreatedAt, &chat.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Chat{}, fmt.Errorf("%w", domain.ErrChatNotFound)
		}

		return domain.Chat{}, fmt.Errorf("an error occurred while updating chat into the database: %v", err)
	}

	chat.ID = dto.ID
	chat.Name = dto.Name
	chat.Description = dto.Description
	chat.CreatorID = dto.CreatorID

	return chat, nil
}

func (r *chatPostgresRepository) Delete(ctx context.Context, memberKey domain.ChatMemberIdentity) error {
	query := "DELETE FROM chats WHERE id = $1 AND creator_id = $2"

	cmgTag, err := r.dbPool.Exec(ctx, query, memberKey.ChatID, memberKey.UserID)
	if err != nil {
		return fmt.Errorf("an error occurred while deleting chat from database: %v", err)
	}

	if cmgTag.RowsAffected() == 0 {
		return fmt.Errorf("%w", domain.ErrChatNotFound)
	}

	return nil
}
