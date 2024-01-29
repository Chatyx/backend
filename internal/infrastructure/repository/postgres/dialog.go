package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/ctxutil"
	"github.com/Chatyx/backend/pkg/log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DialogRepository struct {
	pool   *pgxpool.Pool
	getter dbClientGetter
}

func NewDialogRepository(pool *pgxpool.Pool) *DialogRepository {
	return &DialogRepository{
		pool:   pool,
		getter: dbClientGetter{pool: pool},
	}
}

func (r *DialogRepository) List(ctx context.Context) ([]entity.Dialog, error) {
	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	query := `WITH dialogs AS (
		SELECT	c.id,
				dp.is_blocked,
				c.created_at
		FROM chats c
			INNER JOIN dialog_participants dp
				ON c.id = dp.chat_id
			WHERE dp.user_id = $1
				AND c.type = 'dialog')
	SELECT	dialogs.id,
			dialogs.is_blocked AS "user_is_blocked",
			dp.user_id         AS "partner_user_id",
			dp.is_blocked      AS "partner_is_blocked",
			dialogs.created_at
	FROM dialogs
		INNER JOIN dialog_participants dp
			ON dialogs.id = dp.chat_id
	WHERE dp.user_id != $1`

	rows, err := r.getter.Get(ctx).Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("exec query to select dialogs: %v", err)
	}
	defer rows.Close()

	var dialogs []entity.Dialog

	for rows.Next() {
		var dialog entity.Dialog

		err = rows.Scan(
			&dialog.ID, &dialog.IsBlocked,
			&dialog.Partner.UserID, &dialog.Partner.IsBlocked,
			&dialog.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan dialog row: %v", err)
		}

		dialogs = append(dialogs, dialog)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("reading dialog rows: %v", err)
	}
	return dialogs, nil
}

func (r *DialogRepository) Create(ctx context.Context, dialog *entity.Dialog) error {
	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	partnerUserID := dialog.Partner.UserID
	if userID == partnerUserID {
		return entity.ErrCreateDialogWithYourself
	}

	var (
		hasExternalTx bool
		err           error
	)

	tx, ok := r.getter.Get(ctx).(pgx.Tx)
	if ok {
		hasExternalTx = true
	}

	if !hasExternalTx {
		tx, err = r.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin transaction: %v", err)
		}
		defer func() {
			if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				log.FromContext(ctx).WithError(err).Error("Rollback transaction")
			}
		}()
	}

	query := `INSERT INTO chats
		(uname, type, created_at)
	VALUES ($1, $2, $3)
	RETURNING id`

	err = tx.QueryRow(ctx, query, getUname(userID, partnerUserID), "dialog", dialog.CreatedAt).Scan(&dialog.ID)
	if err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && isDialogUnameUniqueViolation(pgErr) {
			return fmt.Errorf("%w: %s", entity.ErrSuchDialogAlreadyExists, pgErr.Message)
		}

		return fmt.Errorf("exec query to insert dialog: %v", err)
	}

	query = `INSERT INTO dialog_participants
		(chat_id, user_id)
	VALUES ($1, $2), ($1, $3)`

	if _, err = tx.Exec(ctx, query, dialog.ID, userID, partnerUserID); err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && isDialogParticipantUserViolation(pgErr) {
			return fmt.Errorf("%w: %v", entity.ErrCreateDialogWithNonExistentUser, pgErr.Message)
		}
		return fmt.Errorf("exec query to insert dialog participants: %v", err)
	}

	if !hasExternalTx {
		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit transaction: %v", err)
		}
	}

	return nil
}

func (r *DialogRepository) GetByID(ctx context.Context, dialogID int) (entity.Dialog, error) {
	var dialog entity.Dialog

	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	query := `WITH dialogs AS (
		SELECT	c.id,
				dp.is_blocked,
				c.created_at
		FROM chats c
			INNER JOIN dialog_participants dp
				ON c.id = dp.chat_id
			WHERE dp.chat_id = $1
				AND dp.user_id = $2
				AND c.type = 'dialog')
	SELECT	dialogs.id,
			dialogs.is_blocked AS "user_is_blocked",
			dp.user_id         AS "partner_user_id",
			dp.is_blocked      AS "partner_is_blocked",
			dialogs.created_at
	FROM dialogs
		INNER JOIN dialog_participants dp
			ON dialogs.id = dp.chat_id
	WHERE dp.user_id != $2`

	err := r.getter.Get(ctx).QueryRow(ctx, query, dialogID, userID).Scan(
		&dialog.ID, &dialog.IsBlocked,
		&dialog.Partner.UserID, &dialog.Partner.IsBlocked,
		&dialog.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dialog, fmt.Errorf("%w: %v", entity.ErrDialogNotFound, err)
		}

		return dialog, fmt.Errorf("exec query to select dialog: %v", err)
	}

	return dialog, nil
}

func (r *DialogRepository) Update(ctx context.Context, dialog *entity.Dialog) error {
	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	query := `UPDATE dialog_participants
	SET is_blocked = $3
	WHERE chat_id = $1
	  AND user_id != $2`

	execRes, err := r.getter.Get(ctx).Exec(ctx, query, dialog.ID, userID, dialog.Partner.IsBlocked)
	if err != nil {
		return fmt.Errorf("exec query to update dialog participant: %v", err)
	}

	if execRes.RowsAffected() == 0 {
		return fmt.Errorf("%w: there aren't affected rows", entity.ErrDialogNotFound)
	}
	return nil
}

func getUname(ids ...int) string {
	sort.Ints(ids)

	strIds := make([]string, len(ids))
	for i, id := range ids {
		strIds[i] = "u" + strconv.Itoa(id)
	}

	return strings.Join(strIds, ":")
}

func isDialogUnameUniqueViolation(pgErr *pgconn.PgError) bool {
	return pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == "chats_uname_key"
}

func isDialogParticipantUserViolation(pgErr *pgconn.PgError) bool {
	return pgErr.Code == foreignKeyViolationCode && pgErr.ConstraintName == "dialog_participants_user_id_fkey"
}
