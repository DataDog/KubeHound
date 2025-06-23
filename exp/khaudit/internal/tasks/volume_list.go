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

func volumeListCmd() *cobra.Command {
	var params volumeListParams

	cmd := &cobra.Command{
		Use:     "list <run-id>",
		Aliases: []string{"l", "ls"},
		Short:   "list volumes",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeList(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")
	cmd.Flags().StringVar(&params.App, "app", "", "the app to filter by")
	cmd.Flags().StringVar(&params.Team, "team", "", "the team to filter by")
	cmd.Flags().StringVar(&params.Namespace, "namespace", "", "the namespace to filter by")
	cmd.Flags().StringVar(&params.Type, "type", "", "the type to filter by")

	return cmd
}

type volumeListParams struct {
	GremlinEndpoint string
	GremlinAuthMode string

	App       string
	Team      string
	Namespace string
	Type      string
}

func volumeList(ctx context.Context, args []string, params volumeListParams) error {
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
	results, err := volumes.GetVolumes(ctx, args[0], volume.Filter{
		App:       &params.App,
		Team:      &params.Team,
		Namespace: &params.Namespace,
		Type:      &params.Type,
	})
	if err != nil {
		return fmt.Errorf("unable to list volumes: %w", err)
	}

	// Dump as CSV.
	writer := csv.NewWriter(os.Stdout)
	writer.Write([]string{"name", "type", "namespace", "app", "team", "sourcePath", "mountPath", "readOnly"})
	for _, result := range results {
		writer.Write([]string{result.Name, result.Type, result.Namespace, result.App, result.Team, result.SourcePath, result.MountPath, fmt.Sprintf("%v", result.ReadOnly)})
	}

	// Handle errors.
	if err := writer.Error(); err != nil {
		return fmt.Errorf("unable to write CSV: %w", err)
	}

	writer.Flush()

	return nil
}
