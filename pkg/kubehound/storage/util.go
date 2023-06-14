package storage

import (
	"context"
	"time"
)

type Connector[T any] func(ctx context.Context, dbHost string, timeout time.Duration) (T, error)

func Retry[T any](connector Connector[T], retries int, delay time.Duration) Connector[T] {
	return func(ctx context.Context, dbHost string, timeout time.Duration) (T, error) {
		for r := 0; ; r++ {
			var emtpy T
			provider, err := connector(ctx, dbHost, timeout)
			if err == nil || r >= retries {
				return provider, err
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return emtpy, ctx.Err()
			}
		}
	}
}
