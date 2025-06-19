package tasks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/llm"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

type mcpParams struct {
	GremlinEndpoint string
	GremlinAuthMode string
	SSEPort         int16
	SSEPublicHost   string
}

func mcpCmd() *cobra.Command {
	var params mcpParams

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "run a MCP server for AI integrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCP(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")
	cmd.Flags().Int16Var(&params.SSEPort, "sse-port", 0, "MCP HTTP/SSE listening port")
	cmd.Flags().StringVar(&params.SSEPublicHost, "sse-public-host", "locahost", "MCP HTTP/SSE public host")

	return cmd
}

func runMCP(ctx context.Context, _ []string, params mcpParams) error {
	slog.Info("running kubehound MCP server")

	// Initialize gremlin connection.
	conn, err := gremlin.NewConnection(gremlin.Config{
		Endpoint: params.GremlinEndpoint,
		AuthMode: params.GremlinAuthMode,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize gremlin connection: %w", err)
	}

	// Initialize repositories.
	containers := gremlin.Containers(conn)
	ingestions := gremlin.Ingestions(conn)

	// Prepare MCP service.
	mcpServer, err := llm.NewMCPServer(ingestions, containers)
	if err != nil {
		return fmt.Errorf("unable to initialize the MCP service: %w", err)
	}

	var sseServer *server.SSEServer
	if ssePort := params.SSEPort; ssePort > 0 {
		sseServer = mcpServer.ServeSse(params.SSEPublicHost, int(params.SSEPort))
		if err := sseServer.Start(fmt.Sprintf(":%d", ssePort)); err != nil {
			return fmt.Errorf("unable to start MCP HTTP/SSE service: %w", err)
		}
	}
	if err := mcpServer.ServeStdio(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("unable to start MCP Stdio service: %w", err)
	}
	if sseServer != nil {
		_ = sseServer.Shutdown(ctx)
	}

	return nil
}
