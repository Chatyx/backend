package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type chatMemberPostgresRepository struct {
	dbPool PgxPool
	logger logging.Logger
}

func NewChatMemberPostgresRepository(dbPool PgxPool) ChatMemberRepository {
	return &chatMemberPostgresRepository{
		dbPool: dbPool,
		logger: logging.GetLogger(),
	}
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
		r.logger.WithError(err).Error("Unable to list chat members from database")
		return nil, err
	}
	defer rows.Close()

	members := make([]domain.ChatMember, 0)

	for rows.Next() {
		var member domain.ChatMember

		if err = rows.Scan(
			&member.Username, &member.StatusID,
			&member.IsCreator, &member.UserID, &member.ChatID,
		); err != nil {
			r.logger.WithError(err).Error("Unable to scan chat member")
			return nil, err
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while reading chat members")
		return nil, err
	}

	return members, nil
}

func (r *chatMemberPostgresRepository) IsInChat(ctx context.Context, memberKey domain.ChatMemberIdentity) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM chat_members WHERE user_id = $1 AND chat_id = $2 AND status_id = 1)"

	inChat, err := r.exists(ctx, query, memberKey.UserID, memberKey.ChatID)
	if err != nil {
		r.logger.WithError(err).Error("An error occurred while checking if member is in the chat")
		return false, err
	}

	return inChat, nil
}

func (r *chatMemberPostgresRepository) IsChatCreator(ctx context.Context, memberKey domain.ChatMemberIdentity) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM chats WHERE id = $1 AND creator_id = $2)"

	isCreator, err := r.exists(ctx, query, memberKey.ChatID, memberKey.UserID)
	if err != nil {
		r.logger.WithError(err).Error("An error occurred while checking if member is a chat creator")
		return false, err
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
			r.logger.WithError(err).Debug("chat member with such fields is already associated with this chat")
			return domain.ErrChatMemberUniqueViolation
		}

		r.logger.WithError(err).Error("An error occurred while inserting into chat_members table")

		return err
	}

	return nil
}

func (r *chatMemberPostgresRepository) GetByKey(ctx context.Context, memberKey domain.ChatMemberIdentity) (domain.ChatMember, error) {
	logger := r.logger.WithFields(logging.Fields{
		"user_id": memberKey.UserID,
		"chat_id": memberKey.ChatID,
	})
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
			logger.Debug("member is not found")
			return domain.ChatMember{}, domain.ErrChatMemberNotFound
		}

		logger.WithError(err).Error("An error occurred while getting chat member")

		return domain.ChatMember{}, err
	}

	return member, nil
}

func (r *chatMemberPostgresRepository) Update(ctx context.Context, dto domain.UpdateChatMemberDTO) error {
	logger := r.logger.WithFields(logging.Fields{
		"user_id": dto.UserID,
		"chat_id": dto.ChatID,
	})
	query := "UPDATE chat_members SET status_id = $1 WHERE user_id = $2 AND chat_id = $3"

	cmgTag, err := r.dbPool.Exec(ctx, query, dto.StatusID, dto.UserID, dto.ChatID)
	if err != nil {
		logger.WithError(err).Error("An error occurred while updating member")
		return err
	}

	if cmgTag.RowsAffected() == 0 {
		logger.Debugf("member is not found")
		return domain.ErrChatMemberNotFound
	}

	return nil
}
