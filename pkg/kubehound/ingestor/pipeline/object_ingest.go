package pipeline

import (
	"context"
	"reflect"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/hashicorp/go-multierror"
)

// ObjectIngest represents an ingestion pipeline that receives an input object from a collector implementation,
// processes and persists all resulting KubeHound objects (store models, cache entries, graph vertices, etc).
type ObjectIngest interface {
	// Name returns the name of the object ingest pipeline.
	Name() string

	// Initialize intializes an object ingest pipeline with any servic dependencies.
	Initialize(ctx context.Context, deps *Dependencies) error

	// Run executes the ingest pipeline, returning when all are complete.
	Run(ctx context.Context) error

	// Close cleans up any resources held in the ingest pipeline on completion/error.
	Close(ctx context.Context) error
}

// FlushFunc is a callback to be registered in the flush array.
type FlushFunc func(ctx context.Context) (chan struct{}, error)

// CleanupFunc is a callback to be registered in the cleanup array.
type CleanupFunc func(ctx context.Context) error

// objectIngestOptions is a generic container to hold dependencies created on initialization.
// Should not be used directly, but modified via ObjectIngestOption functions.
type objectIngestOptions struct {
	cacheWriter  cache.AsyncWriter                    // Cache provider
	collect      collector.CollectorClient            // Collector fromm which to steam data
	flush        []FlushFunc                          // Array of writer flush functions to be called on a flush
	cleanup      []CleanupFunc                        // Array of dependency cleanup functions to be called on a close
	storeConvert *converter.StoreConverter            // input -> store model converter
	graphConvert *converter.GraphConverter            // store -> graph model converter
	storeWriters map[string]storedb.AsyncWriter       // store writer collection (per model type)
	graphWriters map[string]graphdb.AsyncVertexWriter // graph writer collection (per vertex type)
}

// ObjectIngestOption enables options to be passed to the pipeline initializer.
type ObjectIngestOption func(ctx context.Context, oic *objectIngestOptions, deps *Dependencies) error

// WithCacheWriter initializes a cache writer (and registers a cleanup function) for the ingest pipeline.
func WithCacheWriter() ObjectIngestOption {
	return func(ctx context.Context, oic *objectIngestOptions, deps *Dependencies) error {
		var err error
		oic.cacheWriter, err = deps.Cache.BulkWriter(ctx)
		if err != nil {
			return err
		}

		oic.cleanup = append(oic.cleanup, func(ctx context.Context) error {
			return oic.cacheWriter.Close(ctx)
		})

		oic.flush = append(oic.flush, oic.cacheWriter.Flush)

		return nil
	}
}

// WithCacheWriter initializes a store converter with cache access for the ingest pipeline.
func WithConverterCache() ObjectIngestOption {
	return func(_ context.Context, oic *objectIngestOptions, deps *Dependencies) error {
		oic.storeConvert = converter.NewStoreWithCache(deps.Cache)
		return nil
	}
}

// WithStoreWriter initializes a bulk store writer (and registers a cleanup function) for the provided collection.
// To access the writer use the storeWriter(c collections.Collection) function.
func WithStoreWriter[T collections.Collection](c T) ObjectIngestOption {
	return func(ctx context.Context, oic *objectIngestOptions, deps *Dependencies) error {
		var err error
		w, err := deps.StoreDB.BulkWriter(ctx, c)
		if err != nil {
			return err
		}

		oic.storeWriters[c.Name()] = w
		oic.cleanup = append(oic.cleanup, func(ctx context.Context) error {
			return w.Close(ctx)
		})

		oic.flush = append(oic.flush, w.Flush)

		return nil
	}
}

// WithStoreWriter initializes a bulk graph writer (and registers a cleanup function) for the provided vertex.
// To access the writer use the graphWriter(v vertex.Vertex) function.
func WithGraphWriter[T vertex.Vertex](v T) ObjectIngestOption {
	return func(ctx context.Context, oic *objectIngestOptions, deps *Dependencies) error {
		var err error

		w, err := deps.GraphDB.VertexWriter(ctx, v.Traversal())
		if err != nil {
			return err
		}

		oic.graphWriters[v.Label()] = w
		oic.cleanup = append(oic.cleanup, func(ctx context.Context) error {
			return w.Close(ctx)
		})

		oic.flush = append(oic.flush, w.Flush)

		return nil
	}
}

// BaseObjectIngest provides the base functionality (service initialization, flush and cleanup) for any object ingest pipeline.
type BaseObjectIngest struct {
	opts objectIngestOptions
}

// storeWriter returns the registered store writer for the provided collection.
func (i *BaseObjectIngest) storeWriter(c collections.Collection) storedb.AsyncWriter {
	return i.opts.graphWriters[c.Name()]
}

// graphWriter returns the registered graph writer for the provided collection.
func (i *BaseObjectIngest) graphWriter(v vertex.Vertex) graphdb.AsyncVertexWriter {
	return i.opts.graphWriters[v.Label()]
}

// baseInitialize handles the base initialization of service dependencies for an object ingest pipeline.
// This should be called from the ObjectIngest::Initialize function.
func (i *BaseObjectIngest) baseInitialize(ctx context.Context, deps *Dependencies, opts ...ObjectIngestOption) error {
	var err error

	for _, o := range opts {
		err = o(ctx, &i.opts, deps)
		if err != nil {
			return err
		}
	}

	// If cached option not passed use the default converter
	if i.opts.storeConvert == nil {
		i.opts.storeConvert = converter.NewStore()
	}

	i.opts.graphConvert = converter.NewGraph()

	return nil
}

// cleanup invokes each registered cleanup handler in turn.
// This should be called from the ObjectIngest::Close function.
func (i *BaseObjectIngest) cleanup(ctx context.Context) error {
	var res *multierror.Error

	for _, c := range i.opts.cleanup {
		err := c(ctx)
		if err != nil {
			res = multierror.Append(res, err)
		}
	}

	return res.ErrorOrNil()
}

// flushWriters invokes each registered flush handler in turn and waits for completion.
// This should be called from the ObjectIngest::Flush function.
func (i *BaseObjectIngest) flushWriters(ctx context.Context) error {
	var res *multierror.Error

	waits := make([]chan struct{}, 0)
	for _, flush := range i.opts.flush {
		done, err := flush(ctx)
		if err != nil {
			res = multierror.Append(res, err)
		}

		waits = append(waits, done)
	}

	waitForCompletionMultiple(waits)

	return res.ErrorOrNil()
}

// WaitForCompletionMultiple is a helper function which automatically waits on all the channels and returns the response
// as a slice from each channel. This method is useful to kick off a set of tasks using the WorkerPool and wait for all
// of them to complete before processing each tasks result.
func waitForCompletionMultiple[TOut any](channels []chan TOut) []TOut {
	allResp := make([]TOut, 0, len(channels))
	cases := make([]reflect.SelectCase, len(channels))
	for i, ch := range channels {
		// SelectCase allows us to collect the channel and apply it a specific select case. Therefore, allowing to wait
		// on all these cases simultaneously in the loop below.
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	remaining := len(cases)
	for remaining > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			// The select channel has been closed so zero out the channel to disable the case.
			cases[chosen].Chan = reflect.ValueOf(nil)
			remaining -= 1
			continue
		}

		resp := value.Interface().(TOut)
		allResp = append(allResp, resp)
	}

	return allResp
}
