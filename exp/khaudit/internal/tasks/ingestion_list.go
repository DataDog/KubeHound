package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/ingestion"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/spf13/cobra"
)

func ingestionListCmd() *cobra.Command {
	var params ingestionListParams

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "list ingestions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ingestionList(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")

	return cmd
}

type ingestionListParams struct {
	GremlinEndpoint string
	GremlinAuthMode string
}

func ingestionList(ctx context.Context, _ []string, params ingestionListParams) error {
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

	// List ingestions.
	results, err := ingestions.List(ctx, ingestion.ListFilter{})
	if err != nil {
		return fmt.Errorf("unable to list ingestions: %w", err)
	}

	// Dump as JSON.
	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		return fmt.Errorf("unable to dump as JSON: %w", err)
	}

	return nil
}
