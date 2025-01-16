//go:build no_backend

package main

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/backend"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

func runBackend(ctx context.Context) error {
	l := log.Logger(ctx)

	// Forcing the embed docker config to be loaded
	err := backend.NewBackend(ctx, []string{""}, backend.DefaultUIProfile)
	if err != nil {
		return err
	}
	res, err := backend.IsStackRunning(ctx)
	if err != nil {
		return err
	}
	if !res {
		err = backend.Up(ctx)
		if err != nil {
			return err
		}
	} else {
		l.Info("Backend stack is already running")
	}

	return nil
}

func runBackendCompose(ctx context.Context) error {
	err := backend.NewBackend(ctx, composePath, backend.DefaultUIProfile)
	if err != nil {
		return fmt.Errorf("new backend: %w", err)
	}
	err = backend.Up(ctx)
	if err != nil {
		return fmt.Errorf("docker up: %w", err)
	}

	return nil
}
