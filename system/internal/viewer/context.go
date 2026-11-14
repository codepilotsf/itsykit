// Package viewer carries authenticated request identity, never site configuration.
package viewer

import "context"

type key struct{}
type Admin struct{ Email, CSRF string }

func With(ctx context.Context, admin Admin) context.Context {
	return context.WithValue(ctx, key{}, admin)
}
func From(ctx context.Context) (Admin, bool) { a, ok := ctx.Value(key{}).(Admin); return a, ok }
