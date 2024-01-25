package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/internal/service"
	"github.com/Chatyx/backend/pkg/log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbClient interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type GroupParticipantRepository struct {
	pool *pgxpool.Pool
}

func NewGroupParticipantRepository(pool *pgxpool.Pool) *GroupParticipantRepository {
	return &GroupParticipantRepository{pool: pool}
}

func (r *GroupParticipantRepository) List(ctx context.Context, groupID int) ([]entity.GroupParticipant, error) {
	query := `SELECT gp.chat_id, gp.user_id, gp.status, gp.is_admin
	FROM group_participants gp
	WHERE gp.chat_id = $1`

	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("exec query to select group participants: %v", err)
	}
	defer rows.Close()

	var participants []entity.GroupParticipant

	for rows.Next() {
		var participant entity.GroupParticipant

		err = rows.Scan(
			&participant.GroupID, &participant.UserID,
			&participant.Status, &participant.IsAdmin,
		)
		if err != nil {
			return nil, fmt.Errorf("scan group participant row: %v", err)
		}

		participants = append(participants, participant)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("reading group participant rows: %v", err)
	}
	return participants, nil
}

func (r *GroupParticipantRepository) Get(ctx context.Context, groupID, userID int) (entity.GroupParticipant, error) {
	return r.get(ctx, r.pool, groupID, userID, false)
}

func (r *GroupParticipantRepository) GetThenUpdate(ctx context.Context, groupID, userID int, fn service.GroupParticipantFunc) error {
	if fn == nil {
		if _, err := r.Get(ctx, groupID, userID); err != nil {
			return err
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %v", err)
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.FromContext(ctx).WithError(err).Error("Rollback transaction")
		}
	}()

	participant, err := r.get(ctx, tx, groupID, userID, true)
	if err != nil {
		return err
	}

	if err = fn(&participant); err != nil {
		return err
	}

	if err = r.update(ctx, tx, &participant); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %v", err)
	}
	return nil
}

//nolint:lll // too long namings
func (r *GroupParticipantRepository) get(ctx context.Context, db dbClient, groupID, userID int, forUpdate bool) (entity.GroupParticipant, error) {
	var participant entity.GroupParticipant

	query := `SELECT gp.chat_id, gp.user_id, gp.status, gp.is_admin
	FROM group_participants gp
	WHERE gp.chat_id = $1 AND gp.user_id = $2`

	if forUpdate {
		query += " FOR UPDATE"
	}

	err := db.QueryRow(ctx, query, groupID, userID).Scan(
		&participant.GroupID, &participant.UserID,
		&participant.Status, &participant.IsAdmin,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return participant, fmt.Errorf("%w: %v", entity.ErrGroupParticipantNotFound, err)
		}

		return participant, fmt.Errorf("exec query to select group participant: %v", err)
	}

	return participant, nil
}

func (r *GroupParticipantRepository) update(ctx context.Context, db dbClient, participant *entity.GroupParticipant) error {
	query := "UPDATE group_participants SET status = $3 WHERE chat_id = $1 AND user_id = $2"

	execRes, err := db.Exec(ctx, query, participant.GroupID, participant.UserID, participant.Status)
	if err != nil {
		return fmt.Errorf("exec query to update group participant: %v", err)
	}

	if execRes.RowsAffected() == 0 {
		return fmt.Errorf("%w: there aren't affected rows", entity.ErrGroupParticipantNotFound)
	}
	return nil
}

func (r *GroupParticipantRepository) Create(ctx context.Context, participant *entity.GroupParticipant) error {
	query := "INSERT INTO group_participants (chat_id, user_id, status, is_admin) VALUES ($1, $2, $3, $4)"
	_, err := r.pool.Exec(
		ctx, query,
		participant.GroupID, participant.UserID,
		participant.Status, participant.IsAdmin,
	)
	if err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			switch {
			case isGroupParticipantUserViolation(pgErr):
				return fmt.Errorf("%w: %s", entity.ErrAddNonExistentUserToGroup, pgErr.Message)
			case isGroupParticipantUniqueViolation(pgErr):
				return fmt.Errorf("%w: %s", entity.ErrSuchGroupParticipantAlreadyExists, pgErr.Message)
			}

			return fmt.Errorf("%v", err)
		}

		return fmt.Errorf("exec query to create group participant: %v", err)
	}

	return nil
}

func isGroupParticipantUniqueViolation(pgErr *pgconn.PgError) bool {
	return pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == "group_participants_pkey"
}

func isGroupParticipantUserViolation(pgErr *pgconn.PgError) bool {
	return pgErr.Code == foreignKeyViolationCode && pgErr.ConstraintName == "group_participants_user_id_fkey"
}
