package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/spf13/cobra"
)

func ingestionStatsCmd() *cobra.Command {
	var params ingestionStatsParams

	cmd := &cobra.Command{
		Use:     "stats",
		Aliases: []string{"s"},
		Short:   "ingestion statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ingestionStats(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")

	return cmd
}

type ingestionStatsParams struct {
	GremlinEndpoint string
	GremlinAuthMode string
}

func ingestionStats(ctx context.Context, _ []string, params ingestionStatsParams) error {
	// Initialize gremlin connection.
	conn, err := gremlin.NewConnection(gremlin.Config{
		Endpoint: params.GremlinEndpoint,
		AuthMode: params.GremlinAuthMode,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize gremlin connection: %w", err)
	}

	// Initialize repositories.
	ingestions := gremlin.Ingestions(conn)

	// Get edge counts.
	edgeCounts, err := ingestions.GetEdgeCountPerClasses(ctx)
	if err != nil {
		return fmt.Errorf("unable to get edge count per classes: %w", err)
	}

	// Get vertex counts.
	vertexCounts, err := ingestions.GetVertexCountPerClasses(ctx)
	if err != nil {
		return fmt.Errorf("unable to get vertex count per classes: %w", err)
	}

	// Dump as JSON.
	if err := json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
		"edgeCounts":   edgeCounts,
		"vertexCounts": vertexCounts,
	}); err != nil {
		return fmt.Errorf("unable to dump as JSON: %w", err)
	}

	return nil
}
