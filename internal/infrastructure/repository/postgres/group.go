package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/ctxutil"
	"github.com/Chatyx/backend/pkg/log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepository struct {
	pool *pgxpool.Pool
}

func NewGroupRepository(pool *pgxpool.Pool) *GroupRepository {
	return &GroupRepository{pool: pool}
}

func (r *GroupRepository) List(ctx context.Context) ([]entity.Group, error) {
	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	query := `SELECT c.id,
		c.name,
		c.description,
		c.created_at
	FROM chats c
		INNER JOIN group_participants gp
			ON c.id = gp.chat_id
	WHERE gp.user_id = $1 AND c.type = 'group'`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("exec query to select groups: %v", err)
	}
	defer rows.Close()

	var groups []entity.Group

	for rows.Next() {
		var group entity.Group

		err = rows.Scan(
			&group.ID, &group.Name,
			&group.Description, &group.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan group row: %v", err)
		}

		groups = append(groups, group)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("reading group rows: %v", err)
	}
	return groups, nil
}

func (r *GroupRepository) Create(ctx context.Context, group *entity.Group) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %v", err)
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.FromContext(ctx).WithError(err).Error("Rollback transaction")
		}
	}()

	query := `INSERT INTO chats
		(name, type, description, created_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	err = tx.QueryRow(ctx, query, group.Name, "group", group.Description, group.CreatedAt).Scan(&group.ID)
	if err != nil {
		return fmt.Errorf("exec query to insert group: %v", err)
	}

	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	query = `INSERT INTO group_participants
		(chat_id, user_id, is_admin)
	VALUES ($1, $2, $3)`

	if _, err = tx.Exec(ctx, query, group.ID, userID, true); err != nil {
		return fmt.Errorf("exec query to insert group participant: %v", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %v", err)
	}

	return nil
}

func (r *GroupRepository) GetByID(ctx context.Context, groupID int) (entity.Group, error) {
	var group entity.Group

	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	query := `SELECT c.id,
		   c.name,
		   c.description,
		   c.created_at
	FROM chats c
			 INNER JOIN group_participants gp
						ON c.id = gp.chat_id
	WHERE c.id = $1
	  AND c.type = 'group'
	  AND gp.user_id = $2`

	err := r.pool.QueryRow(ctx, query, groupID, userID).Scan(
		&group.ID, &group.Name,
		&group.Description, &group.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return group, fmt.Errorf("%v: %w", err, entity.ErrGroupNotFound)
		}

		return group, fmt.Errorf("exec query to select group: %v", err)
	}

	return group, nil
}

func (r *GroupRepository) Update(ctx context.Context, group *entity.Group) error {
	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	query := `UPDATE chats AS c
	SET name        = $3,
		description = $4,
		updated_at  = $5
	FROM group_participants AS gp
	WHERE c.id = gp.chat_id
	  AND c.id = $1
	  AND c.type = 'group'
	  AND gp.user_id = $2
	  AND gp.is_admin IS TRUE
	RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		group.ID, userID,
		group.Name, group.Description, time.Now(),
	).Scan(&group.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%v: %w", err, entity.ErrGroupNotFound)
		}
		return fmt.Errorf("exec query to update group: %v", err)
	}

	return nil
}

func (r *GroupRepository) Delete(ctx context.Context, groupID int) error {
	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	query := `DELETE
	FROM chats AS c
		USING group_participants gp
	WHERE c.id = gp.chat_id
	  AND c.id = $1
	  AND c.type = 'group'
	  AND gp.user_id = $2
	  AND gp.is_admin IS TRUE`

	execRes, err := r.pool.Exec(ctx, query, groupID, userID)
	if err != nil {
		return fmt.Errorf("exec query to delete group: %v", err)
	}

	if execRes.RowsAffected() == 0 {
		return fmt.Errorf("there's no affected rows: %w", entity.ErrGroupNotFound)
	}
	return nil
}
