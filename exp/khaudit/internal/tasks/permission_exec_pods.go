package tasks

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/spf13/cobra"
)

type permissionExecPodsParams struct {
	GremlinEndpoint string
	GremlinAuthMode string
}

func permissionExecPodsCmd() *cobra.Command {
	var params permissionExecPodsParams

	cmd := &cobra.Command{
		Use:     "exec-pods <run-id> <group-name>",
		Long:    "Kubectl executable pods are pods that can be executed using a kubectl exec.",
		Aliases: []string{"ep"},
		Short:   "permission exec pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			return permissionExecPods(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")

	return cmd
}

func permissionExecPods(ctx context.Context, args []string, params permissionExecPodsParams) error {
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

	// Get enterable pods.
	execPods, err := permissions.GetKubectlExecutablePodCount(ctx, args[0], args[1])
	if err != nil {
		return fmt.Errorf("unable to get enterable pods: %w", err)
	}

	// Print enterable pods.
	fmt.Println(execPods)

	return nil
}
