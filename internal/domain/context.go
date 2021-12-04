package domain

import (
	"context"
)

type userCtxKey struct{}

type userIDCtxKey struct{}

var (
	userKey   = userCtxKey{}
	userIDKey = userIDCtxKey{}
)

func NewContextFromUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}

	return ""
}

func NewContextFromAuthUser(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func AuthUserFromContext(ctx context.Context) AuthUser {
	if user, ok := ctx.Value(userKey).(AuthUser); ok {
		return user
	}

	return AuthUser{}
}
