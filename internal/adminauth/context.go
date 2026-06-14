package adminauth

import "context"

type ctxKey string

const adminCtxKey ctxKey = "admin_auth"

func WithAdminAuth(ctx context.Context, claims *AdminClaims) context.Context {
	return context.WithValue(ctx, adminCtxKey, claims)
}

func GetAdminAuth(ctx context.Context) *AdminClaims {
	claims, _ := ctx.Value(adminCtxKey).(*AdminClaims)
	return claims
}
