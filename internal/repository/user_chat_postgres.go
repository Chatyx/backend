package repository

import (
	"context"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/jackc/pgtype"
)

type userChatPostgresRepository struct {
	dbPool PgxPool
	logger logging.Logger
}

func NewUserChatPostgresRepository(dbPool PgxPool) UserChatRepository {
	return &userChatPostgresRepository{
		dbPool: dbPool,
		logger: logging.GetLogger(),
	}
}

func (r *userChatPostgresRepository) ListUsersWhoBelongToChat(ctx context.Context, chatID string) ([]domain.User, error) {
	query := `SELECT 
		users.id, users.username, users.password, 
		users.first_name, users.last_name, users.email, 
		users.birth_date, users.department, users.is_deleted, 
		users.created_at, users.updated_at
	FROM users
	INNER JOIN users_chats ON users.id = users_chats.user_id
	WHERE users_chats.chat_id = $1`

	rows, err := r.dbPool.Query(ctx, query, chatID)
	if err != nil {
		r.logger.WithError(err).Error("Unable to list users from database")
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)

	for rows.Next() {
		var (
			user      domain.User
			birthDate pgtype.Date
		)

		if err = rows.Scan(
			&user.ID, &user.Username, &user.Password,
			&user.FirstName, &user.LastName, &user.Email,
			&birthDate, &user.Department, &user.IsDeleted,
			&user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			r.logger.WithError(err).Error("Unable to scan user")
			return nil, err
		}

		if birthDate.Status == pgtype.Present {
			user.BirthDate = birthDate.Time.Format("2006-01-02")
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while reading users")
		return nil, err
	}

	return users, nil
}

func (r *userChatPostgresRepository) IsUserBelongToChat(ctx context.Context, userID, chatID string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users_chats WHERE user_id = $1 AND chat_id = $2)"

	isBelong := false
	row := r.dbPool.QueryRow(ctx, query, userID, chatID)

	if err := row.Scan(&isBelong); err != nil {
		r.logger.WithError(err).Error("An error occurred while checking if user belongs to the chat")
		return false, nil
	}

	return isBelong, nil
}
