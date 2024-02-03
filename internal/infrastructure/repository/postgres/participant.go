package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Chatyx/backend/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupParticipantRepository struct {
	pool   *pgxpool.Pool
	getter dbClientGetter
}

func NewGroupParticipantRepository(pool *pgxpool.Pool) *GroupParticipantRepository {
	return &GroupParticipantRepository{
		pool:   pool,
		getter: dbClientGetter{pool: pool},
	}
}

func (r *GroupParticipantRepository) List(ctx context.Context, groupID int) ([]entity.GroupParticipant, error) {
	query := `SELECT gp.chat_id, gp.user_id, gp.status, gp.is_admin
	FROM group_participants gp
	WHERE gp.chat_id = $1`

	rows, err := r.getter.Get(ctx).Query(ctx, query, groupID)
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

func (r *GroupParticipantRepository) Get(ctx context.Context, groupID, userID int, withLock bool) (entity.GroupParticipant, error) {
	var participant entity.GroupParticipant

	query := `SELECT gp.chat_id, gp.user_id, gp.status, gp.is_admin
	FROM group_participants gp
	WHERE gp.chat_id = $1 AND gp.user_id = $2`

	if withLock {
		query += " FOR UPDATE"
	}

	err := r.getter.Get(ctx).QueryRow(ctx, query, groupID, userID).Scan(
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

func (r *GroupParticipantRepository) Update(ctx context.Context, participant *entity.GroupParticipant) error {
	query := "UPDATE group_participants SET status = $3 WHERE chat_id = $1 AND user_id = $2"

	execRes, err := r.getter.Get(ctx).Exec(ctx, query, participant.GroupID, participant.UserID, participant.Status)
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
	_, err := r.getter.Get(ctx).Exec(
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
