package repository

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

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
