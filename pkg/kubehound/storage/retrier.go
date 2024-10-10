package storage

import (
	"context"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
)

type Connector[T any] func(ctx context.Context, cfg *config.KubehoundConfig) (T, error)

func Retrier[T any](connector Connector[T], retries int, delay time.Duration) Connector[T] {
	return func(ctx context.Context, cfg *config.KubehoundConfig) (T, error) {
		for r := 0; ; r++ {
			var empty T
			provider, err := connector(ctx, cfg)
			if err == nil || r >= retries {
				return provider, err
			}
			//log.I..Warnf("Retrying to connect [%d/%d]", r+1, retries)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return empty, ctx.Err()
			}
		}
	}
}
