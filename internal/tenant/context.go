package tenant

import "context"

type ctxKey string

const tenantCtxKey ctxKey = "tenant_id"

func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantCtxKey, tenantID)
}

func GetTenant(ctx context.Context) string {
	id, _ := ctx.Value(tenantCtxKey).(string)
	return id
}
