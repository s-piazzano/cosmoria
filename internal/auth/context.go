package auth

import "context"

type ctxKey string

const claimsKey ctxKey = "auth_claims"

func WithAuth(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func GetAuth(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsKey).(*Claims)
	return claims
}
