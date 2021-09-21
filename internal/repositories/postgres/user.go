package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/utils"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepository struct {
	dbPool *pgxpool.Pool

	logger logging.Logger
}

func NewUserRepository(dbPool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		dbPool: dbPool,
		logger: logging.GetLogger(),
	}
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	query := `
		SELECT 
			id, username, password, 
			first_name, last_name, email, 
			birth_date, department, is_deleted, 
			created_at, updated_at
		FROM users
	`

	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("Unable to list users from database")

		return nil, err
	}
	defer rows.Close()

	users := make([]*domain.User, 0)

	for rows.Next() {
		user := new(domain.User)

		var (
			birthDate sql.NullTime
			updatedAt sql.NullTime
		)

		if err = rows.Scan(
			&user.ID, &user.Username, &user.Password,
			&user.FirstName, &user.LastName, &user.Email,
			&birthDate, &user.Department, &user.IsDeleted,
			&user.CreatedAt, &updatedAt,
		); err != nil {
			r.logger.WithError(err).Error("Unable to scan user")

			return nil, err
		}

		user.BirthDate = birthDate.Time
		user.UpdatedAt = updatedAt.Time

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error occurred while reading users")

		return nil, err
	}

	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, dto domain.CreateUserDTO) (*domain.User, error) {
	id := ""
	query := `
		INSERT INTO users(
			username, password, first_name, 
			last_name, email, birth_date, department
		) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`

	if err := r.dbPool.QueryRow(
		ctx, query,
		dto.Username, dto.Password, dto.FirstName,
		dto.LastName, dto.Email, dto.BirthDate, dto.Department,
	).Scan(&id); err != nil {
		r.logger.WithError(err).Error("Error occurred while creating user into the database")

		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if !utils.IsValidUUID(id) {
		return nil, fmt.Errorf("user is not found")
	}

	return r.getBy(ctx, "id", id)
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.getBy(ctx, "username", username)
}

func (r *UserRepository) getBy(ctx context.Context, fieldName string, fieldValue interface{}) (*domain.User, error) {
	query := fmt.Sprintf(`
		SELECT 
			id, username, password, 
			first_name, last_name, email, 
			birth_date, department, is_deleted, 
			created_at, updated_at
		FROM users WHERE %s = $1`,
		fieldName,
	)

	user := new(domain.User)

	var (
		birthDate sql.NullTime
		updatedAt sql.NullTime
	)

	row := r.dbPool.QueryRow(ctx, query, fieldValue)
	if err := row.Scan(
		&user.ID, &user.Username, &user.Password,
		&user.FirstName, &user.LastName, &user.Email,
		&birthDate, &user.Department, &user.IsDeleted,
		&user.CreatedAt, &updatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user is not found")
		}

		r.logger.WithError(err).Errorf("Error occurred while getting user by %s", fieldName)

		return nil, nil
	}

	user.BirthDate = birthDate.Time
	user.UpdatedAt = updatedAt.Time

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET
			username = $2, password = $3, 
			first_name = $4, last_name = $5, email = $6, 
			birth_date = $7, department = $8, is_deleted = $9, 
			created_at = $10, updated_at = $11
		WHERE id = $1 
	`

	if _, err := r.dbPool.Exec(
		ctx, query,
		user.ID,
		user.Username, user.Password,
		user.FirstName, user.LastName, user.Email,
		user.BirthDate, user.Department, user.IsDeleted,
		user.CreatedAt, user.UpdatedAt,
	); err != nil {
		r.logger.WithError(err).Error("Error occurred while updating user into the database")

		return err
	}

	return nil
}
