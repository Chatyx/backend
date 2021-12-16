package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

type chatMemberPostgresRepository struct {
	dbPool PgxPool
}

func NewChatMemberPostgresRepository(dbPool PgxPool) ChatMemberRepository {
	return &chatMemberPostgresRepository{dbPool: dbPool}
}

func (r *chatMemberPostgresRepository) ListByChatID(ctx context.Context, chatID string) ([]domain.ChatMember, error) {
	query := `SELECT users.username, chat_members.status_id, 
		chat_members.user_id = chats.creator_id, chat_members.user_id, chat_members.chat_id
	FROM chat_members 
	INNER JOIN users 
		ON users.id = chat_members.user_id
	INNER JOIN chats
		ON chats.id = chat_members.chat_id
	WHERE chat_members.chat_id = $1`

	return r.list(ctx, query, chatID)
}

func (r *chatMemberPostgresRepository) ListByUserID(ctx context.Context, userID string) ([]domain.ChatMember, error) {
	query := `SELECT users.username, chat_members.status_id, 
		chat_members.user_id = chats.creator_id, chat_members.user_id, chat_members.chat_id
	FROM chat_members 
	INNER JOIN users 
		ON users.id = chat_members.user_id
	INNER JOIN chats
		ON chats.id = chat_members.chat_id
	WHERE chat_members.user_id = $1`

	return r.list(ctx, query, userID)
}

func (r *chatMemberPostgresRepository) list(ctx context.Context, query string, args ...interface{}) ([]domain.ChatMember, error) {
	rows, err := r.dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("an error occured while quering list of chat members from database: %v", err)
	}
	defer rows.Close()

	members := make([]domain.ChatMember, 0)

	for rows.Next() {
		var member domain.ChatMember

		if err = rows.Scan(
			&member.Username, &member.StatusID,
			&member.IsCreator, &member.UserID, &member.ChatID,
		); err != nil {
			return nil, fmt.Errorf("an error occurred while scanning chat member: %v", err)
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occurred while reading chat members: %v", err)
	}

	return members, nil
}

func (r *chatMemberPostgresRepository) IsInChat(ctx context.Context, memberKey domain.ChatMemberIdentity) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM chat_members WHERE user_id = $1 AND chat_id = $2 AND status_id = 1)"

	inChat, err := r.exists(ctx, query, memberKey.UserID, memberKey.ChatID)
	if err != nil {
		return false, fmt.Errorf("an error occurred while checking if member is in the chat: %v", err)
	}

	return inChat, nil
}

func (r *chatMemberPostgresRepository) IsChatCreator(ctx context.Context, memberKey domain.ChatMemberIdentity) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM chats WHERE id = $1 AND creator_id = $2)"

	isCreator, err := r.exists(ctx, query, memberKey.ChatID, memberKey.UserID)
	if err != nil {
		return false, fmt.Errorf("an error occurred while checking if member is a chat creator: %v", err)
	}

	return isCreator, nil
}

func (r *chatMemberPostgresRepository) exists(ctx context.Context, query string, args ...interface{}) (bool, error) {
	var result bool

	row := r.dbPool.QueryRow(ctx, query, args...)
	if err := row.Scan(&result); err != nil {
		return false, err
	}

	return result, nil
}

func (r *chatMemberPostgresRepository) Create(ctx context.Context, memberKey domain.ChatMemberIdentity) error {
	query := "INSERT INTO chat_members (user_id, chat_id) VALUES ($1, $2)"

	if _, err := r.dbPool.Exec(ctx, query, memberKey.UserID, memberKey.ChatID); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("%w", domain.ErrChatMemberUniqueViolation)
		}

		return fmt.Errorf("an error occurred while inserting chat member into the database: %v", err)
	}

	return nil
}

func (r *chatMemberPostgresRepository) GetByKey(ctx context.Context, memberKey domain.ChatMemberIdentity) (domain.ChatMember, error) {
	query := `SELECT users.username, chat_members.status_id, 
		chat_members.user_id = chats.creator_id, chat_members.user_id, chat_members.chat_id
	FROM chat_members 
	INNER JOIN users 
		ON users.id = chat_members.user_id
	INNER JOIN chats
		ON chats.id = chat_members.chat_id
	WHERE chat_members.user_id = $1 AND chat_members.chat_id = $2`

	var member domain.ChatMember

	if err := r.dbPool.QueryRow(ctx, query, memberKey.UserID, memberKey.ChatID).Scan(
		&member.Username, &member.StatusID,
		&member.IsCreator, &member.UserID, &member.ChatID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ChatMember{}, fmt.Errorf("%w", domain.ErrChatMemberNotFound)
		}

		return domain.ChatMember{}, fmt.Errorf("an error occurred while getting chat member: %v", err)
	}

	return member, nil
}

func (r *chatMemberPostgresRepository) Update(ctx context.Context, dto domain.UpdateChatMemberDTO) error {
	query := "UPDATE chat_members SET status_id = $1 WHERE user_id = $2 AND chat_id = $3"

	cmgTag, err := r.dbPool.Exec(ctx, query, dto.StatusID, dto.UserID, dto.ChatID)
	if err != nil {
		return fmt.Errorf("an error occurred while updating member: %v", err)
	}

	if cmgTag.RowsAffected() == 0 {
		return fmt.Errorf("%w", domain.ErrChatMemberNotFound)
	}

	return nil
}
