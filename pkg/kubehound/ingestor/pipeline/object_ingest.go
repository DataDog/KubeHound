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

type ObjectIngest interface {
	Name() string
	Initialize(ctx context.Context, deps *Dependencies) error // Any errors abort everything
	Run(ctx context.Context) error                            // Any errors abort everything
	Close(ctx context.Context) error
}

type FlushFunc func(ctx context.Context) (chan struct{}, error)
type CleanupFunc func(ctx context.Context) error

type objectIngestOptions struct {
	cacheWriter  cache.AsyncWriter
	ingest       collector.CollectorClient
	flush        []FlushFunc
	cleanup      []CleanupFunc
	storeConvert *converter.StoreConverter
	graphConvert *converter.GraphConverter
	storeWriters map[string]storedb.AsyncWriter
	graphWriters map[string]graphdb.AsyncVertexWriter
}

type ObjectIngestOption func(ctx context.Context, oic *objectIngestOptions, deps *Dependencies) error

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

func WithConverterCache() ObjectIngestOption {
	return func(_ context.Context, oic *objectIngestOptions, deps *Dependencies) error {
		oic.storeConvert = converter.NewStoreWithCache(deps.Cache)
		return nil
	}
}

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

type BaseObjectIngest struct {
	opts objectIngestOptions
}

func (i *BaseObjectIngest) storeWriter(c collections.Collection) storedb.AsyncWriter {
	return i.opts.graphWriters[c.Name()]
}

func (i *BaseObjectIngest) graphWriter(v vertex.Vertex) graphdb.AsyncVertexWriter {
	return i.opts.graphWriters[v.Label()]
}

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
