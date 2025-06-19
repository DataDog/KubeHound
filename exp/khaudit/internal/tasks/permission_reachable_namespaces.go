package tasks

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sort"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/spf13/cobra"
)

type permissionReachableNamespacesParams struct {
	GremlinEndpoint string
	GremlinAuthMode string
}

func permissionReachableNamespacesCmd() *cobra.Command {
	var params permissionReachableNamespacesParams

	cmd := &cobra.Command{
		Use:     "reachable-namespaces <run-id> <group-name>",
		Long:    "Reachable namespaces are namespaces containing pods that can be reached from a kubectl exec.",
		Aliases: []string{"rn"},
		Short:   "permission reachable namespaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			return permissionReachableNamespaces(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")

	return cmd
}

func permissionReachableNamespaces(ctx context.Context, args []string, params permissionReachableNamespacesParams) error {
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

	// Get reachable namespaces.
	exposedPodCounts, err := permissions.GetExposedPodCountPerNamespace(ctx, args[0], args[1])
	if err != nil {
		return fmt.Errorf("unable to get reachable namespaces: %w", err)
	}

	// Sort the exposed pod counts by pod count.
	sort.Slice(exposedPodCounts, func(i, j int) bool {
		return exposedPodCounts[i].PodCount > exposedPodCounts[j].PodCount
	})

	// Dump as CSV.
	writer := csv.NewWriter(os.Stdout)
	writer.Write([]string{"group_name", "namespace", "pod_count"})

	// Write each exposed pod count.
	for _, exposedPodCount := range exposedPodCounts {
		writer.Write([]string{exposedPodCount.GroupName, exposedPodCount.Namespace, fmt.Sprintf("%d", exposedPodCount.PodCount)})
	}

	// Flush the writer.
	writer.Flush()

	return nil
}
