package tasks

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/container"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/ingestion"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type containerEscapeContainersParams struct {
	Cluster string
	RunID   string

	GremlinEndpoint string
	GremlinAuthMode string
}

// containerEscapeContainersWorkUnit represents a container escape path impacted containers work unit.
type containerEscapeContainersWorkUnit struct {
	run       ingestion.Ingestion
	namespace string
}

func containerEscapeContainersCmd() *cobra.Command {
	var params containerEscapeProfilesParams

	cmd := &cobra.Command{
		Use:     "containers",
		Aliases: []string{"c"},
		Short:   "List container escape path impacted containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return containerEscapeContainers(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.Cluster, "cluster", "", "the cluster to filter by")
	cmd.Flags().StringVar(&params.RunID, "run-id", "", "the run identifier to filter by")
	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")

	return cmd
}

func containerEscapeContainers(ctx context.Context, args []string, params containerEscapeProfilesParams) error {
	slog.Info("running khaudit")

	// Check arguments.
	if len(args) == 0 {
		return fmt.Errorf("missing namespace argument")
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
	runChan := make(chan containerEscapeContainersWorkUnit, 10)
	resultChan := make(chan container.Container, 1000)

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
		// Create a CSV writer.
		csvWriter := csv.NewWriter(os.Stdout)
		defer csvWriter.Flush()

		// Write the header.
		if err := csvWriter.Write([]string{"cluster", "run_id", "namespace", "app", "team", "image"}); err != nil {
			return fmt.Errorf("unable to write header: %w", err)
		}

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

				// Write the record.
				if err := csvWriter.Write([]string{c.Cluster, c.RunID, c.Namespace, c.App, c.Team, c.Image}); err != nil {
					return fmt.Errorf("unable to write csv record: %w", err)
				}

				// Flush the CSV writer.
				csvWriter.Flush()
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

				// Enumerate containers.
				if err := containers.GetVulnerables(egCtx, w.run.Cluster, w.run.RunID, container.AttackPathFilter{
					Namespace: &w.namespace,
				}, resultChan); err != nil {
					if errors.Is(err, container.ErrNoResult) {
						slog.Info("no vulnerable containers found", "namespace", w.namespace)
						continue
					}

					return fmt.Errorf("unable to get vulnerable containers: %w", err)
				}
			}
		}
	})

	// Send runs to the run channel.
	for _, ns := range args {
		for _, work := range runs {
			runChan <- containerEscapeContainersWorkUnit{
				run:       work,
				namespace: ns,
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
