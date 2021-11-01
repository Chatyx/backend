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

func (r *chatMemberPostgresRepository) ListMembersInChat(ctx context.Context, chatID string) ([]domain.ChatMember, error) {
	query := `SELECT users.username, chat_members.status_id, 
		chat_members.user_id, chat_members.chat_id
	FROM users 
	INNER JOIN chat_members 
		ON users.id = chat_members.user_id
	WHERE chat_members.chat_id = $1 AND chat_members.status_id = 1`

	rows, err := r.dbPool.Query(ctx, query, chatID)
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
			&member.UserID, &member.ChatID,
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

func (r *chatMemberPostgresRepository) IsMemberInChat(ctx context.Context, userID, chatID string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM chat_members WHERE user_id = $1 AND chat_id = $2 AND status_id = 1)"

	isIn := false
	row := r.dbPool.QueryRow(ctx, query, userID, chatID)

	if err := row.Scan(&isIn); err != nil {
		r.logger.WithError(err).Error("An error occurred while checking if member is in the chat")
		return false, err
	}

	return isIn, nil
}

func (r *chatMemberPostgresRepository) Create(ctx context.Context, userID, chatID string) error {
	query := "INSERT INTO chat_members (user_id, chat_id) VALUES ($1, $2)"

	if _, err := r.dbPool.Exec(ctx, query, userID, chatID); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			r.logger.WithError(err).Debug("chat member with such fields is already associated with this chat")
			return domain.ErrChatMemberUniqueViolation
		}

		r.logger.WithError(err).Error("An error occurred while inserting into chat_members table")

		return err
	}

	return nil
}

func (r *chatMemberPostgresRepository) Get(ctx context.Context, userID, chatID string) (domain.ChatMember, error) {
	logger := r.logger.WithFields(logging.Fields{
		"user_id": userID,
		"chat_id": chatID,
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

	if err := r.dbPool.QueryRow(ctx, query, userID, chatID).Scan(
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
