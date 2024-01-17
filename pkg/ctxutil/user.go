package ctxutil

import "context"

type userKey struct{}

func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userKey{}).(string); ok {
		return id
	}

	return ""
}

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userKey{}, id)
}
