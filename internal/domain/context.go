package domain

import "context"

type userCtxKey struct{}

var userKey = userCtxKey{}

func NewContextFromUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(userKey).(*User); ok {
		return user
	}

	return nil
}
