package main

import (
	"log/slog"
	"os"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/tasks"
)

func main() {
	// Initialize logger.
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With(slog.String("service", "khaudit"))
	slog.SetDefault(logger)

	// Run the command.
	if err := run(); err != nil {
		slog.Error("unable to run command", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Execute the root command.
	return tasks.Execute()
}
