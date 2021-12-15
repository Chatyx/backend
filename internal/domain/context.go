package domain

import (
	"context"
)

type userCtxKey struct{}

var userKey = userCtxKey{}

func NewContextFromAuthUser(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func AuthUserFromContext(ctx context.Context) AuthUser {
	if user, ok := ctx.Value(userKey).(AuthUser); ok {
		return user
	}

	return AuthUser{}
}
