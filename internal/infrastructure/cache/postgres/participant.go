package postgres

import (
	"context"
	"fmt"

	"github.com/Chatyx/backend/internal/entity"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type ParticipantChecker struct {
	pool *pgxpool.Pool
}

func NewParticipantChecker(pool *pgxpool.Pool) *ParticipantChecker {
	return &ParticipantChecker{pool: pool}
}

func (c ParticipantChecker) Check(ctx context.Context, chatID entity.ChatID, userID int) error {
	b := builder.Select("1").
		Prefix("SELECT EXISTS (").
		Suffix(")")
	if chatID.Type == entity.GroupChatType {
		b = b.From("group_participants").Where(sq.Eq{"status": entity.JoinedStatus})
	} else {
		b = b.From("dialog_participants").Where(sq.Eq{"is_blocked": false})
	}

	query, args, err := b.Where(sq.Eq{"chat_id": chatID.ID}).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query to check existance of participant: %v", err)
	}

	var exist bool
	if err = c.pool.QueryRow(ctx, query, args...).Scan(&exist); err != nil {
		return fmt.Errorf("scan existance of participant result: %v", err)
	}

	if !exist {
		if chatID.Type == entity.GroupChatType {
			return entity.ErrGroupNotFound
		}
		return entity.ErrDialogNotFound
	}
	return nil
}
