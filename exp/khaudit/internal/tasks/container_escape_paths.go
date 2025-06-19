package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/container"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/ingestion"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/kubehound"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type containerEscapePathsParams struct {
	Cluster   string
	RunID     string
	Namespace string

	GremlinEndpoint string
	GremlinAuthMode string

	Images []string
}

// containerEscapePathWorkUnit represents a container escape path impacted containers work unit.
type containerEscapePathsWorkUnit struct {
	run ingestion.Ingestion

	image     *string
	namespace string
}

func containerEscapePathsCmd() *cobra.Command {
	var params containerEscapePathsParams

	cmd := &cobra.Command{
		Use:     "paths",
		Aliases: []string{"p"},
		Short:   "List container escape paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			return containerEscapePaths(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringSliceVar(&params.Images, "images", nil, "the images to filter by")
	cmd.Flags().StringVar(&params.Cluster, "cluster", "", "the cluster to filter by")
	cmd.Flags().StringVar(&params.RunID, "run-id", "", "the run identifier to filter by")
	cmd.Flags().StringVar(&params.Namespace, "namespace", "", "the namespace to filter by")
	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")

	return cmd
}

func containerEscapePaths(ctx context.Context, args []string, params containerEscapePathsParams) error {
	slog.Info("running khaudit")

	// Check arguments.
	if params.Namespace == "" {
		return fmt.Errorf("missing namespace argument")
	}
	if params.RunID == "" {
		return fmt.Errorf("missing run-id argument")
	}
	if params.Cluster == "" {
		return fmt.Errorf("missing cluster argument")
	}
	if len(args) > 0 {
		slog.Warn("deprecated: this function does not expect any arguments. Use the --images flag instead.")
		params.Images = append(params.Images, args...)
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
	containers := gremlin.Containers(conn)
	ingestions := gremlin.Ingestions(conn)

	// Count by namespaces.
	runChan := make(chan containerEscapePathsWorkUnit, 10)
	resultChan := make(chan kubehound.AttackPath, 1000)

	// Enumerate ingestions.
	runs, err := ingestions.List(ctx, ingestion.ListFilter{
		Cluster: &params.Cluster,
		RunID:   &params.RunID,
	})
	if err != nil {
		return fmt.Errorf("unable to list ingestions: %w", err)
	}
	if len(runs) == 0 {
		slog.Warn("no ingestions found")
		return nil
	}

	// Syncronisation error group.
	eg, egCtx := errgroup.WithContext(ctx)

	// Container escape path result handler.
	eg.Go(func() error {
		// Loop over the result channel.
		for {
			select {
			case <-egCtx.Done():
				return ctx.Err()
			case c, ok := <-resultChan:
				if !ok {
					// Channel closed.
					return nil
				}

				json.NewEncoder(os.Stdout).Encode(c)
			}
		}
	})

	// Query for containers with attack paths.
	eg.Go(func() error {
		defer close(resultChan)

		for {
			select {
			case <-egCtx.Done():
				return ctx.Err()
			case w, ok := <-runChan:
				if !ok {
					// Channel closed.
					return nil
				}

				slog.Info("processing cluster ingested data", "cluster", w.run.Cluster, "run_id", w.run.RunID, "namespace", w.namespace)

				// Enumerate attack paths for the namespace and image.
				filter := container.AttackPathFilter{
					Namespace: &w.namespace,
				}
				if w.image != nil {
					filter.Image = w.image
				}
				if err := containers.GetAttackPaths(egCtx, w.run.Cluster, w.run.RunID, filter, resultChan); err != nil {
					if errors.Is(err, container.ErrNoResult) {
						logInfo := []any{"namespace", w.namespace} // Can't use a []string because slog expect a []any
						if w.image != nil {
							logInfo = append(logInfo, "image", *w.image)
						}
						slog.Info("no attack found", logInfo...)
						continue
					}

					return fmt.Errorf("unable to get attack paths: %w", err)
				}
			}
		}
	})

	// Send runs to the run channel.
	for _, work := range runs {
		switch {
		case len(params.Images) > 0:
			for _, img := range params.Images {
				runChan <- containerEscapePathsWorkUnit{
					run:       work,
					namespace: params.Namespace,
					image:     &img,
				}
			}
		// If we ever add other filters - container name, team, etc, add case here
		default:
			runChan <- containerEscapePathsWorkUnit{
				run:       work,
				namespace: params.Namespace,
			}
		}
	}

	// Close the run channel.
	close(runChan)

	// Wait for the error group.
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("unable to wait for error group: %w", err)
	}

	return nil
}
