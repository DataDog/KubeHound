package pipeline

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/hashicorp/go-multierror"
)

// FlushFunc is a callback to be registered in the flush array.
type FlushFunc func(ctx context.Context) error

// CleanupFunc is a callback to be registered in the cleanup array.
type CleanupFunc func(ctx context.Context) error

// resourceOptions is a generic container to hold dependencies created on initialization.
// Should not be used directly, but modified via ObjectIngestOption functions.
type resourceOptions struct {
	db           *sql.DB                            // SQLite database handle
	storeDB      storedb.Provider                   // Store DB provider for direct writes
	collect      collector.CollectorClient           // Collector from which to stream data
	flush        []FlushFunc                         // Array of writer flush functions to be called on a flush
	cleanup      []CleanupFunc                       // Array of dependency cleanup functions to be called on a close
	storeConvert *converter.StoreConverter           // input -> store model converter
	graphConvert *converter.GraphConverter           // store -> graph model converter
	graphWriters map[string]graphdb.AsyncVertexWriter // graph writer collection (per vertex type)
}

// IngestResourceOption enables options to be passed to the pipeline initializer.
type IngestResourceOption func(ctx context.Context, oic *resourceOptions, deps *Dependencies) error

// WithConverterDB initializes a store converter with database access for the ingest pipeline.
func WithConverterDB() IngestResourceOption {
	return func(_ context.Context, rOpts *resourceOptions, deps *Dependencies) error {
		db, ok := deps.StoreDB.Reader().(*sql.DB)
		if !ok {
			return fmt.Errorf("store provider reader is not *sql.DB")
		}
		rOpts.db = db
		rOpts.storeConvert = converter.NewStoreWithDB(deps.Config, db)

		return nil
	}
}

// WithGraphWriter initializes a bulk graph writer (and registers a cleanup function) for the provided vertex.
// To access the writer use the graphWriter(v vertex.Vertex) function.
func WithGraphWriter(v vertex.Builder) IngestResourceOption {
	return func(ctx context.Context, rOpts *resourceOptions, deps *Dependencies) error {
		if err := v.Initialize(deps.Config); err != nil {
			return err
		}

		db, ok := deps.StoreDB.Reader().(*sql.DB)
		if !ok {
			return fmt.Errorf("store provider reader is not *sql.DB")
		}

		w, err := deps.GraphDB.VertexWriter(ctx, v, db, graphdb.WithTags(tag.GetBaseTags()))
		if err != nil {
			return err
		}

		rOpts.graphWriters[v.Label()] = w
		rOpts.cleanup = append(rOpts.cleanup, func(ctx context.Context) error {
			return w.Close(ctx)
		})

		rOpts.flush = append(rOpts.flush, w.Flush)

		return nil
	}
}

// IngestResources provides the base functionality (service initialization, flush and cleanup) for any object ingest pipeline.
type IngestResources struct {
	resourceOptions
}

// writeStore delegates a write to the store DB provider.
func (i *IngestResources) writeStore(ctx context.Context, model any) error {
	return i.storeDB.Write(ctx, model)
}

// writeVertex delegates a write to the registered graph writer after invoking the vertex.Processor on the provided insert.
func (i *IngestResources) writeVertex(ctx context.Context, v vertex.Builder, insert any) error {
	w := i.graphWriters[v.Label()]

	processed, err := v.Processor(ctx, insert)
	if err != nil {
		return fmt.Errorf("vertex processing: %w", err)
	}

	return w.Queue(ctx, processed)
}

// identityExists checks if an identity with the given name and namespace already exists in the database.
func (i *IngestResources) identityExists(ctx context.Context, name, namespace string) bool {
	if i.db == nil {
		return false
	}
	var id int64
	err := i.db.QueryRowContext(ctx,
		"SELECT id FROM identities WHERE name = ? AND namespace = ? LIMIT 1",
		name, namespace).Scan(&id)
	return err == nil
}

// endpointSliceExists checks if an endpoint slice exists for the given parameters.
func (i *IngestResources) endpointSliceExists(ctx context.Context, namespace, podName, protocol string, port int) bool {
	if i.db == nil {
		return false
	}
	var id int64
	err := i.db.QueryRowContext(ctx,
		"SELECT id FROM endpoints WHERE namespace = ? AND pod_name = ? AND protocol = ? AND port = ? AND has_slice = 1 LIMIT 1",
		namespace, podName, protocol, port).Scan(&id)
	return err == nil
}

// CreateResources handles the base initialization of service dependencies for an object ingest pipeline.
// This should be called from the ObjectIngest::Initialize function.
func CreateResources(ctx context.Context, deps *Dependencies, opts ...IngestResourceOption) (*IngestResources, error) {
	var err error

	i := &IngestResources{
		resourceOptions{
			collect:      deps.Collector,
			storeDB:      deps.StoreDB,
			graphConvert: converter.NewGraph(deps.Config),
			storeConvert: converter.NewStore(deps.Config),
			flush:        make([]FlushFunc, 0),
			cleanup:      make([]CleanupFunc, 0),
			graphWriters: make(map[string]graphdb.AsyncVertexWriter),
		},
	}

	// Do a cleanup of whatever has been registered in the case of a partial success
	defer func() {
		if err != nil {
			err := i.cleanupAll(ctx)
			if err != nil {
				log.Trace(ctx).Errorf("Ingestor cleanup failure: %v", err)
			}
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
