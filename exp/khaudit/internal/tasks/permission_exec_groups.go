package tasks

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/spf13/cobra"
)

type permissionExecGroupsParams struct {
	GremlinEndpoint string
	GremlinAuthMode string
}

func permissionExecGroupsCmd() *cobra.Command {
	var params permissionExecGroupsParams

	cmd := &cobra.Command{
		Use:     "exec-groups <run-id> <namespace>",
		Long:    "Kubectl executable groups are groups that have kubectl executable pods in a namespace.",
		Aliases: []string{"eg"},
		Short:   "permission exec groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			return permissionExecGroups(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")

	return cmd
}

func permissionExecGroups(ctx context.Context, args []string, params permissionExecGroupsParams) error {
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
	permissions := gremlin.Permissions(conn)

	// Get reachable pods.
	groups, err := permissions.GetKubectlExecutableGroupsForNamespace(ctx, args[0], args[1])
	if err != nil {
		return fmt.Errorf("unable to get reachable pods: %w", err)
	}

	// Dump as CSV.
	writer := csv.NewWriter(os.Stdout)
	writer.Write([]string{"namespace", "group_name"})

	// Write each group.
	for _, group := range groups {
		writer.Write([]string{args[1], group})
	}

	// Flush the writer.
	writer.Flush()

	return nil
}
