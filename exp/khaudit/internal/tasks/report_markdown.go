package tasks

import (
	"context"
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/renderer"
	"github.com/spf13/cobra"
)

type reportMarkdownParams struct {
	GremlinEndpoint string
	GremlinAuthMode string
	Namespaces      []string
	AssumeGroup     string
}

func reportMarkdownRootCmd() *cobra.Command {
	var params reportMarkdownParams

	cmd := &cobra.Command{
		Use:     "markdown <cluster> <run-id>",
		Aliases: []string{"md"},
		Short:   "markdown report generation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return reportMarkdown(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")
	cmd.Flags().StringSliceVar(&params.Namespaces, "namespaces", []string{}, "the namespaces to include in the report")
	cmd.Flags().StringVar(&params.AssumeGroup, "assume-group", "employees", "the group to assume")

	return cmd
}

func reportMarkdown(ctx context.Context, args []string, params reportMarkdownParams) error {
	// Check arguments.
	if len(args) != 2 {
		return fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

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
	permissions := gremlin.Permissions(conn)
	containers := gremlin.Containers(conn)
	volumes := gremlin.Volumes(conn)

	// Initialize renderer.
	mdRenderer := renderer.Markdown(ingestions, permissions, containers, volumes, params.Namespaces, params.AssumeGroup)

	// Render report.
	if err := mdRenderer.Render(ctx, os.Stdout, args[0], args[1]); err != nil {
		return fmt.Errorf("unable to render report: %w", err)
	}

	return nil
}
