package ingestor

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/pipeline"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
)

type PipelineIngestor struct {
	cfg       *config.KubehoundConfig
	collector collector.CollectorClient
	cache     cache.CacheProvider
	storedb   storedb.Provider
	graphdb   graphdb.Provider
	sequences []pipeline.Sequence
}

// TODO dependency inject cache, store DB, graph DB
// TODO health check all dependencies
// TODO run all components in order. Each component shpould execute the appropriate pipeline

func New(cfg *config.KubehoundConfig, collect collector.CollectorClient, cache cache.CacheProvider, storedb storedb.Provider,
	graphdb graphdb.Provider) (Ingestor, error) {

	n := &PipelineIngestor{
		cfg:       cfg,
		collector: collect,
		cache:     cache,
		storedb:   storedb,
		graphdb:   graphdb,
		sequences: make([]pipeline.Sequence, 0),
	}

	n.sequences = append(n.sequences)
	// [roles, clusterroles] -> [rolebinding, clusterrolebdinog]
	// [nodes] -> [pods]

	return n, nil
}

func (i PipelineIngestor) handleHealthCheck(ok bool, err error) error {
	return nil
}

func (i PipelineIngestor) HealthCheck(ctx context.Context) error {
	// Check ingestor
	i.handleHealthCheck(i.collector.HealthCheck(ctx))

	// Check cache connection
	// Check store connection
	// Check graph connection

	return nil
}

func (i PipelineIngestor) Run(ctx context.Context) error {

	// Call the ingestor to stream the components in order
	// Pipelines
	wg := &sync.WaitGroup{}
	for _, seq := range i.sequences {
		s := seq
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := s.Run(ctx, &pipeline.Dependencies{
				Config:  i.cfg,
				Cache:   i.cache,
				StoreDB: i.storedb,
				GraphDB: i.graphdb,
			}); err != nil {
				// TODO cancel everything on error
			}
		}()
	}

	wg.Wait()

	return nil
}

func (i PipelineIngestor) Close(ctx context.Context) error {
	return nil
}
