package tasks

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/container"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/ingestion"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/infrastructure/persistence/gremlin"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const PROFILE_STEP_SEPARATOR = "-->"

type containerEscapeProfilesParams struct {
	Cluster string
	RunID   string

	GremlinEndpoint string
	GremlinAuthMode string

	Namespaces         []string
	ExcludedNamespaces []string
}

func containerEscapeProfilesCmd() *cobra.Command {
	var params containerEscapeProfilesParams

	cmd := &cobra.Command{
		Use:     "profiles",
		Aliases: []string{"p"},
		Short:   "List container escape path profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			return containerEscapeProfiles(cmd.Context(), args, params)
		},
	}

	cmd.Flags().StringVar(&params.Cluster, "cluster", "", "the cluster to filter by")
	cmd.Flags().StringVar(&params.RunID, "run-id", "", "the run identifier to filter by")
	cmd.Flags().StringVar(&params.GremlinEndpoint, "gremlin-endpoint", "ws://localhost:8182/gremlin", "the gremlin endpoint")
	cmd.Flags().StringVar(&params.GremlinAuthMode, "gremlin-auth-mode", "plain", "the gremlin auth mode")
	cmd.Flags().StringSliceVar(&params.Namespaces, "namespaces", nil, "the namespaces to filter by")
	cmd.Flags().StringSliceVar(&params.ExcludedNamespaces, "excluded-namespaces", nil, "the excluded namespaces")

	return cmd
}

func containerEscapeProfiles(ctx context.Context, _ []string, params containerEscapeProfilesParams) error {
	slog.Info("running khaudit")

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
	runChan := make(chan ingestion.Ingestion, 1)
	resultChan := make(chan container.NamespaceAggregation, 1000)

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

	// Namespace aggregation handler.
	eg.Go(func() error {
		// Create a CSV writer.
		csvWriter := csv.NewWriter(os.Stdout)
		defer csvWriter.Flush()

		// Write the header.
		if err := csvWriter.Write([]string{"cluster", "run_id", "namespace", "pod_count", "attack", "attack_path", "path_count"}); err != nil {
			return fmt.Errorf("unable to write header: %w", err)
		}

		// Loop over the result channel.
		for {
			select {
			case <-egCtx.Done():
				return ctx.Err()
			case ns, ok := <-resultChan:
				if !ok {
					// Channel closed.
					return nil
				}

				slog.Info("namespace aggregation", "namespace", ns.Namespace, "pod_count", ns.Count)

				// Skip empty namespaces.
				if ns.Count == 0 {
					// No containers in the namespace.
					slog.Warn("no containers in the namespace", "namespace", ns.Namespace)
					continue
				}

				// Enumerate attach paths for the namespace.
				paths, err := containers.GetAttackPathProfiles(egCtx, ns.Cluster, ns.RunID, container.AttackPathFilter{
					Namespace: &ns.Namespace,
				})
				if err != nil {
					if errors.Is(err, container.ErrNoResult) {
						slog.Info("no attack path profiles found", "namespace", ns.Namespace)
						continue
					}

					return fmt.Errorf("unable to get attack path profiles: %w", err)
				}

				// Log attack paths.
				for _, path := range paths {
					// Write the record.
					if err := csvWriter.Write([]string{ns.Cluster, ns.RunID, ns.Namespace, fmt.Sprintf("%d", ns.Count), "container-escape", strings.Join(path.Path, PROFILE_STEP_SEPARATOR), fmt.Sprintf("%d", path.ContainerCount)}); err != nil {
						return fmt.Errorf("unable to write csv record: %w", err)
					}
				}

				// Flush the CSV writer.
				csvWriter.Flush()
			}
		}
	})

	// Count by namespaces handler.
	eg.Go(func() error {
		defer close(resultChan)

		for {
			select {
			case <-egCtx.Done():
				return ctx.Err()
			case run, ok := <-runChan:
				if !ok {
					// Channel closed.
					return nil
				}

				slog.Info("processing cluster ingested data", "cluster", run.Cluster, "run_id", run.RunID)

				// Count by namespaces.
				if err := containers.CountByNamespaces(egCtx, run.Cluster, run.RunID, container.NamespaceAggregationFilter{
					Namespaces:         params.Namespaces,
					ExcludedNamespaces: params.ExcludedNamespaces,
				}, resultChan); err != nil {
					return fmt.Errorf("unable to count by namespaces: %w", err)
				}
			}
		}
	})

	// Send runs to the run channel.
	for _, run := range runs {
		runChan <- run
	}

	// Close the run channel.
	close(runChan)

	// Wait for the error group.
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("unable to wait for error group: %w", err)
	}

	return nil
}
