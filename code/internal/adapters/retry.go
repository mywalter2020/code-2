package adapters

import (
	"context"
	"time"
)

func retryDo[T any](ctx context.Context, attempts int, delay time.Duration, fn func() (T, error)) (T, error) {
	var zero T
	if attempts <= 1 {
		return fn()
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		out, err := fn()
		if err == nil {
			return out, nil
		}
		lastErr = err
		if i == attempts-1 {
			break
		}
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
		}
	}
	return zero, lastErr
}
