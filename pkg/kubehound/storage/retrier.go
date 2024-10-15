package storage

import (
	"context"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

type Connector[T any] func(ctx context.Context, cfg *config.KubehoundConfig) (T, error)

func Retrier[T any](connector Connector[T], retries int, delay time.Duration) Connector[T] {
	return func(ctx context.Context, cfg *config.KubehoundConfig) (T, error) {
		l := log.Logger(ctx)
		for r := 0; ; r++ {
			var empty T
			provider, err := connector(ctx, cfg)
			if err == nil || r >= retries {
				return provider, err
			}
			l.Warn("Retrying to connect", log.Int("attempt", r+1), log.Int("retries", retries))

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return empty, ctx.Err()
			}
		}
	}
}
