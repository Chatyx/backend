package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Chatyx/backend/pkg/log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

type dbClient interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type dbClientGetter struct {
	pool *pgxpool.Pool
}

func (g dbClientGetter) Get(ctx context.Context) dbClient { //nolint:ireturn // it's necessary to return abstraction for this case
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return g.pool
}

type TransactionManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

func (txm *TransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := txm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %v", err)
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.FromContext(ctx).WithError(err).Error("Rollback transaction")
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err = fn(txCtx); err != nil {
		return fmt.Errorf("call inner func: %v", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %v", err)
	}
	return nil
}
