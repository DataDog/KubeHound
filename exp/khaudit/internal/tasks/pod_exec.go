package tasks

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/permission"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/spf13/cobra"
)

type podExecParams struct {
	GremlinEndpoint string
	GremlinAuthMode string

	Image string
	App   string
	Team  string
}

func podExecRootCmd() *cobra.Command {
	var params podExecParams

	cmd := &cobra.Command{
		Use:     "exec <run-id> <namespace> <group-name>",
		Aliases: []string{"e"},
		Long:    "List the pods that can be executed using kubectl exec in a namespace for a group.",
		Short:   "pod exec",
		RunE: func(cmd *cobra.Command, args []string) error {
			return podExec(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")
	cmd.Flags().StringVar(&params.Image, "image", "", "the image to filter by")
	cmd.Flags().StringVar(&params.App, "app", "", "the app to filter by")
	cmd.Flags().StringVar(&params.Team, "team", "", "the team to filter by")

	return cmd
}

func podExec(ctx context.Context, args []string, params podExecParams) error {
	// Check arguments.
	if len(args) != 3 {
		return fmt.Errorf("expected 3 arguments, got %d", len(args))
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

	// Get the pods that can be executed using kubectl exec in the namespace for the group.
	pods, err := permissions.GetExposedNamespacePods(ctx, args[0], args[1], args[2], permission.ExposedPodFilter{
		Image: &params.Image,
		App:   &params.App,
		Team:  &params.Team,
	})
	if err != nil {
		return fmt.Errorf("unable to get pods: %w", err)
	}

	// Dump the pods as CSV.
	writer := csv.NewWriter(os.Stdout)
	writer.Write([]string{"namespace", "pod", "image", "app", "team"})
	for _, pod := range pods {
		writer.Write([]string{pod.Namespace, pod.PodName, pod.Image, pod.App, pod.Team})
	}

	writer.Flush()

	return nil
}
