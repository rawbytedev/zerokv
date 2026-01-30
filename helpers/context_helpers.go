package helpers

import "context"

// IgnoreContext runs the provided function while respecting the context's cancellation.
func IgnoreContext(ctx context.Context, fn func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn()
}

type result struct {
	data []byte
	err  error
}

// RunWithContext runs the provided function while respecting the context's cancellation.
func RunWithContext(ctx context.Context, fn func() ([]byte, error)) ([]byte, error) {
	done := make(chan result, 1)
	go func() {
		data, err := fn()
		done <- result{data, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-done:
		return res.data, res.err
	}
}
