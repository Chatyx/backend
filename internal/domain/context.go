package domain

import "context"

type userCtxKey struct{}

var userKey = userCtxKey{}

func NewContextFromUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userKey, userID)
}

func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userKey).(string); ok {
		return id
	}

	return ""
}
