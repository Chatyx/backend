package ctxutil

import (
	"context"
	"strconv"
)

type UserID string

func (uid UserID) ToInt() int {
	res, err := strconv.Atoi(string(uid))
	if err != nil {
		return 0
	}

	return res
}

type userKey struct{}

func UserIDFromContext(ctx context.Context) UserID {
	if id, ok := ctx.Value(userKey{}).(UserID); ok {
		return id
	}

	return ""
}

func WithUserID(ctx context.Context, uid UserID) context.Context {
	return context.WithValue(ctx, userKey{}, uid)
}
