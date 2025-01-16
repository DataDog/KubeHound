//go:build !no_backend

package main

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

func runBackend(ctx context.Context) error {
	l := log.Logger(ctx)
	l.Warn("Backend is not supported in this build")

	return nil
}

func runBackendCompose(ctx context.Context) error {
	l := log.Logger(ctx)
	l.Warn("Backend is not supported in this build")

	return nil
}
