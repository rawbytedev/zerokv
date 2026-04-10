package internal

import "context"

// CheckContext returns the context error if the context is cancelled, otherwise nil.
func CheckContext(ctx context.Context) error {
	return ctx.Err()
}