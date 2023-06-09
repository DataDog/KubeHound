package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/hashicorp/go-multierror"
)

// FlushFunc is a callback to be registered in the flush array.
type FlushFunc func(ctx context.Context) error

// CleanupFunc is a callback to be registered in the cleanup array.
type CleanupFunc func(ctx context.Context) error

// resourceOptions is a generic container to hold dependencies created on initialization.
// Should not be used directly, but modified via ObjectIngestOption functions.
type resourceOptions struct {
	cacheWriter  cache.AsyncWriter                    // Cache provider
	collect      collector.CollectorClient            // Collector fromm which to steam data
	flush        []FlushFunc                          // Array of writer flush functions to be called on a flush
	cleanup      []CleanupFunc                        // Array of dependency cleanup functions to be called on a close
	storeConvert *converter.StoreConverter            // input -> store model converter
	graphConvert *converter.GraphConverter            // store -> graph model converter
	storeWriters map[string]storedb.AsyncWriter       // store writer collection (per model type)
	graphWriters map[string]graphdb.AsyncVertexWriter // graph writer collection (per vertex type)
}

// IngestResourceOption enables options to be passed to the pipeline initializer.
type IngestResourceOption func(ctx context.Context, oic *resourceOptions, deps *Dependencies) error

// WithCacheWriter initializes a cache writer (and registers a cleanup function) for the ingest pipeline.
func WithCacheWriter() IngestResourceOption {
	return func(ctx context.Context, rOpts *resourceOptions, deps *Dependencies) error {
		var err error
		rOpts.cacheWriter, err = deps.Cache.BulkWriter(ctx)
		if err != nil {
			return err
		}

		rOpts.cleanup = append(rOpts.cleanup, func(ctx context.Context) error {
			return rOpts.cacheWriter.Close(ctx)
		})

		rOpts.flush = append(rOpts.flush, rOpts.cacheWriter.Flush)

		return nil
	}
}

// WithCacheWriter initializes a store converter with cache access for the ingest pipeline.
func WithConverterCache() IngestResourceOption {
	return func(_ context.Context, rOpts *resourceOptions, deps *Dependencies) error {
		rOpts.storeConvert = converter.NewStoreWithCache(deps.Cache)
		return nil
	}
}

// WithStoreWriter initializes a bulk store writer (and registers a cleanup function) for the provided collection.
// To access the writer use the storeWriter(c collections.Collection) function.
func WithStoreWriter[T collections.Collection](c T) IngestResourceOption {
	return func(ctx context.Context, rOpts *resourceOptions, deps *Dependencies) error {
		w, err := deps.StoreDB.BulkWriter(ctx, c)
		if err != nil {
			return err
		}

		rOpts.storeWriters[c.Name()] = w
		rOpts.cleanup = append(rOpts.cleanup, func(ctx context.Context) error {
			return w.Close(ctx)
		})

		rOpts.flush = append(rOpts.flush, w.Flush)

		return nil
	}
}

// WithStoreWriter initializes a bulk graph writer (and registers a cleanup function) for the provided vertex.
// To access the writer use the graphWriter(v vertex.Vertex) function.
func WithGraphWriter(v vertex.Builder) IngestResourceOption {
	log.I.Infof("--- WithGraphWriter: %+v", v)
	return func(ctx context.Context, rOpts *resourceOptions, deps *Dependencies) error {
		log.I.Infof("--- callback func for WithGraphWriter: %+v", v)
		w, err := deps.GraphDB.VertexWriter(ctx, v)
		if err != nil {
			return err
		}
		log.I.Infof("--- deps.GraphDB.VertexWriter : %+v", v)

		rOpts.graphWriters[v.Label()] = w
		rOpts.cleanup = append(rOpts.cleanup, func(ctx context.Context) error {
			return w.Close(ctx)
		})

		log.I.Infof("--- append cleanup ok : %+v", v)
		rOpts.flush = append(rOpts.flush, w.Flush)
		log.I.Infof("--- all append done : %+v", v)

		return nil
	}
}

// IngestResources provides the base functionality (service initialization, flush and cleanup) for any object ingest pipeline.
type IngestResources struct {
	resourceOptions
}

// storeWriter returns the registered store writer for the provided collection.
func (i *IngestResources) storeWriter(c collections.Collection) storedb.AsyncWriter {
	return i.storeWriters[c.Name()]
}

// graphWriter returns the registered graph writer for the provided collection.
func (i *IngestResources) graphWriter(v vertex.Builder) graphdb.AsyncVertexWriter {
	return i.graphWriters[v.Label()]
}

// CreateResources handles the base initialization of service dependencies for an object ingest pipeline.
// This should be called from the ObjectIngest::Initialize function.
func CreateResources(ctx context.Context, deps *Dependencies, opts ...IngestResourceOption) (*IngestResources, error) {
	var err error

	i := &IngestResources{
		resourceOptions{
			collect:      deps.Collector,
			graphConvert: converter.NewGraph(),
			storeConvert: converter.NewStore(),
			flush:        make([]FlushFunc, 0),
			cleanup:      make([]CleanupFunc, 0),
			graphWriters: make(map[string]graphdb.AsyncVertexWriter),
			storeWriters: make(map[string]storedb.AsyncWriter),
		},
	}

	// Do a cleanup of whatever has been registered in the case of a partial success
	defer func() {
		if err != nil {
			i.cleanupAll(ctx)
		}
	}()

	for _, o := range opts {
		err = o(ctx, &i.resourceOptions, deps)
		if err != nil {
			return nil, err
		}
	}

	return i, nil
}

// cleanup invokes each registered cleanup handler in turn.
// This should be called from the ObjectIngest::Close function.
func (i *IngestResources) cleanupAll(ctx context.Context) error {
	var res *multierror.Error

	for _, c := range i.cleanup {
		err := c(ctx)
		if err != nil {
			res = multierror.Append(res, err)
		}
	}

	// Empty the cleanup to ensure it is only called once
	i.cleanup = make([]CleanupFunc, 0)

	return res.ErrorOrNil()
}

// flushWriters invokes each registered flush handler in turn and waits for completion.
// This should be called from the ObjectIngest::Flush function.
func (i *IngestResources) flushWriters(ctx context.Context) error {
	var res *multierror.Error

	for _, flush := range i.flush {
		if err := flush(ctx); err != nil {
			res = multierror.Append(res, err)
		}
	}

	return res.ErrorOrNil()
}
