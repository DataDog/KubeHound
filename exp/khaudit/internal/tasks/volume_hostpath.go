package tasks

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/volume"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/spf13/cobra"
)

func volumeHostpathCmd() *cobra.Command {
	var params volumeHostpathParams

	cmd := &cobra.Command{
		Use:     "hostpath <run-id>",
		Aliases: []string{"hp"},
		Short:   "list hostpath volumes",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeHostpath(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")
	cmd.Flags().StringVar(&params.SourcePath, "source-path", "", "the source path to filter by")
	cmd.Flags().StringVar(&params.Namespace, "namespace", "", "the namespace to filter by")
	cmd.Flags().StringVar(&params.Image, "image", "", "the image to filter by")
	cmd.Flags().StringVar(&params.App, "app", "", "the app to filter by")
	cmd.Flags().StringVar(&params.Team, "team", "", "the team to filter by")

	return cmd
}

type volumeHostpathParams struct {
	GremlinEndpoint string
	GremlinAuthMode string

	// Filters.
	SourcePath string
	Namespace  string
	Image      string
	App        string
	Team       string
}

func volumeHostpath(ctx context.Context, args []string, params volumeHostpathParams) error {
	// Initialize gremlin connection.
	conn, err := gremlin.NewConnection(gremlin.Config{
		Endpoint: params.GremlinEndpoint,
		AuthMode: params.GremlinAuthMode,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize gremlin connection: %w", err)
	}

	// Initialize repositories.
	volumes := gremlin.Volumes(conn)

	// List volumes.
	results, err := volumes.GetMountedHostPaths(ctx, args[0], volume.Filter{
		Namespace:  &params.Namespace,
		Image:      &params.Image,
		App:        &params.App,
		Team:       &params.Team,
		SourcePath: &params.SourcePath,
	})
	if err != nil {
		return fmt.Errorf("unable to list volumes: %w", err)
	}

	// Dump as CSV.
	writer := csv.NewWriter(os.Stdout)
	writer.Write([]string{"sourcePath", "namespace", "app", "team", "image"})
	for _, result := range results {
		writer.Write([]string{result.SourcePath, result.Namespace, result.App, result.Team, result.Image})
	}

	// Handle errors.
	if err := writer.Error(); err != nil {
		return fmt.Errorf("unable to write CSV: %w", err)
	}

	writer.Flush()

	return nil
}
