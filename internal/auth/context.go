package auth

import "context"

type contextKey string

const userContextKey contextKey = "authUser"

type AuthUser struct {
	ID    uint
	Email string
	Role  string
}

func WithUser(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (AuthUser, bool) {
	user, ok := ctx.Value(userContextKey).(AuthUser)
	return user, ok
}

func UserIDString(ctx context.Context) string {
	if user, ok := UserFromContext(ctx); ok {
		return user.Email
	}
	return "system"
}
