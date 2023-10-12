package ingestor

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/pipeline"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// PipelineIngestor is a parallelized pipeline based ingestor implementation.
type PipelineIngestor struct {
	cfg       *config.KubehoundConfig
	collector collector.CollectorClient
	cache     cache.CacheProvider
	storedb   storedb.Provider
	graphdb   graphdb.Provider
	sequences []pipeline.Sequence
}

// ingestSequence returns the optimized pipeline sequence for ingestion.
func ingestSequence() []pipeline.Sequence {
	return []pipeline.Sequence{
		{
			Name: "core-pipeline",
			Groups: []pipeline.Group{
				{
					Name: "k8s-role-group",
					Ingests: []pipeline.ObjectIngest{
						&pipeline.RoleIngest{},
						&pipeline.ClusterRoleIngest{},
					},
				},
				{
					Name: "k8s-binding-group",
					Ingests: []pipeline.ObjectIngest{
						&pipeline.RoleBindingIngest{},
						&pipeline.ClusterRoleBindingIngest{},
					},
				},
				{
					Name: "k8s-core-group",
					Ingests: []pipeline.ObjectIngest{
						&pipeline.NodeIngest{},
						&pipeline.EndpointIngest{},
					},
				},
				{
					Name: "k8s-pod-group",
					Ingests: []pipeline.ObjectIngest{
						&pipeline.PodIngest{},
					},
				},
			},
		},
	}
}

// newPipelineIngestor creates a new pipeline ingestor instance.
func newPipelineIngestor(cfg *config.KubehoundConfig, collect collector.CollectorClient, c cache.CacheProvider,
	storedb storedb.Provider, graphdb graphdb.Provider) (Ingestor, error) {

	n := &PipelineIngestor{
		cfg:       cfg,
		collector: collect,
		cache:     c,
		storedb:   storedb,
		graphdb:   graphdb,
		sequences: ingestSequence(),
	}

	return n, nil
}

// HealthCheck enables a check of the ingestor service dependencies.
func (i PipelineIngestor) HealthCheck(ctx context.Context) error {
	return services.HealthCheck(ctx, []services.Dependency{
		i.cache,
		i.storedb,
		i.graphdb,
		i.collector,
	})
}

// Run executes the pipeline ingest and blocks until complete.
func (i PipelineIngestor) Run(outer context.Context) error {
	ctx, cancelAll := context.WithCancelCause(outer)
	defer cancelAll(nil)

	l := log.Trace(ctx, log.WithComponent(globals.IngestorComponent))
	l.Info("Starting ingest sequences")

	wg := &sync.WaitGroup{}
	deps := &pipeline.Dependencies{
		Config:    i.cfg,
		Collector: i.collector,
		Cache:     i.cache,
		StoreDB:   i.storedb,
		GraphDB:   i.graphdb,
	}

	// Run the sequences in parallel and cancel ingest on any errors. Note we deliberately avoid
	// using a worker pool here as have a small, fixed number of tasks to run in parallel.
	for _, seq := range i.sequences {
		s := seq
		wg.Add(1)

		go func() {
			defer wg.Done()
			l.Infof("Running ingestor sequence %s", s.Name)

			err := s.Run(ctx, deps)
			if err != nil {
				l.Errorf("ingestor sequence %s run: %v", s.Name, err)
				cancelAll(err)
			}
		}()
	}

	l.Info("Waiting for ingest sequences to complete")
	wg.Wait()

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	l.Info("Completed pipeline ingest")

	return nil
}

// Close cleans up any resources owned by the pipeline ingestor.
func (i PipelineIngestor) Close(ctx context.Context) error {
	// No ownership of dependencies. Nothing to do here
	return nil
}
