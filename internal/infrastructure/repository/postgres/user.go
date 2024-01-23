package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Chatyx/backend/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) List(ctx context.Context) ([]entity.User, error) {
	query := `SELECT id,
		username,
		pwd_hash,
		email,
		first_name,
		last_name,
		birth_date,
		bio,
		created_at
	FROM users
	WHERE deleted_at IS NULL`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("exec query to select users: %v", err)
	}
	defer rows.Close()

	var users []entity.User

	for rows.Next() {
		var user entity.User

		err = rows.Scan(
			&user.ID, &user.Username, &user.PwdHash,
			&user.Email, &user.FirstName, &user.LastName,
			&user.BirthDate, &user.Bio, &user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user row: %v", err)
		}

		users = append(users, user)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("reading user rows: %v", err)
	}
	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (
    	username, pwd_hash, email,
		first_name, last_name, birth_date,
		bio, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id`

	err := r.pool.QueryRow(ctx, query,
		user.Username, user.PwdHash, user.Email,
		user.FirstName, user.LastName, user.BirthDate,
		user.Bio, user.CreatedAt,
	).Scan(&user.ID)
	if err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return fmt.Errorf("%w: %s", entity.ErrSuchUserAlreadyExists, pgErr.Message)
		}

		return fmt.Errorf("exec query to insert user: %v", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (entity.User, error) {
	query := `SELECT id,
		username,
		pwd_hash,
		email,
		first_name,
		last_name,
		birth_date,
		bio,
		created_at
	FROM users
	WHERE id = $1
	  AND deleted_at IS NULL`

	return r.getBy(ctx, query, id)
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	query := `SELECT id,
		username,
		pwd_hash,
		email,
		first_name,
		last_name,
		birth_date,
		bio,
		created_at
	FROM users
	WHERE username = $1
	  AND deleted_at IS NULL`

	return r.getBy(ctx, query, username)
}

func (r *UserRepository) getBy(ctx context.Context, query string, args ...any) (entity.User, error) {
	var user entity.User

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Username, &user.PwdHash,
		&user.Email, &user.FirstName, &user.LastName,
		&user.BirthDate, &user.Bio, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, fmt.Errorf("%w: %v", entity.ErrUserNotFound, err)
		}

		return user, fmt.Errorf("exec query to select user: %v", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `UPDATE users
	SET	username   = $2,
		email      = $3,
		first_name = $4,
		last_name  = $5,
		birth_date = $6,
		bio        = $7,
		updated_at = $8
	WHERE id = $1
	  AND deleted_at IS NULL
	RETURNING pwd_hash, created_at`

	err := r.pool.QueryRow(ctx, query, user.ID,
		user.Username, user.Email, user.FirstName,
		user.LastName, user.BirthDate,
		user.Bio, time.Now(),
	).Scan(&user.PwdHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: %v", entity.ErrUserNotFound, err)
		}
		return fmt.Errorf("exec query to update user: %v", err)
	}

	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID int, pwdHash string) error {
	query := `UPDATE users
	SET	pwd_hash   = $2,
		updated_at = $3
	WHERE id = $1
	  AND deleted_at IS NULL`

	execRes, err := r.pool.Exec(ctx, query, userID, pwdHash, time.Now())
	if err != nil {
		return fmt.Errorf("exec query to update user password: %v", err)
	}

	if execRes.RowsAffected() == 0 {
		return fmt.Errorf("%w: there aren't affected rows", entity.ErrUserNotFound)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE users
	SET	updated_at = $2,
		deleted_at = $2
	WHERE id = $1
	  AND deleted_at IS NULL`

	execRes, err := r.pool.Exec(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("exec query to update user deleted_at: %v", err)
	}

	if execRes.RowsAffected() == 0 {
		return fmt.Errorf("%w: there aren't affected rows", entity.ErrUserNotFound)
	}
	return nil
}
