package graphdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	channelSizeBatchFactor = 4 // TODO maybe move that into a config file?
	StorageProviderName    = "janusgraph"
)

var (
	_ Provider = (*JanusGraphProvider)(nil)
)

type JanusGraphProvider struct {
	drc  *gremlin.DriverRemoteConnection // Connection to the remote gremlin server
	tags []string                        // Tags to be applied for telemetry
	cfg  *config.KubehoundConfig         // Application configuration
}

func NewGraphDriver(ctx context.Context, cfg *config.KubehoundConfig) (*JanusGraphProvider, error) {
	if cfg.JanusGraph.URL == "" {
		return nil, errors.New("JanusGraph DB URL is not set")
	}

	driver, err := gremlin.NewDriverRemoteConnection(cfg.JanusGraph.URL,
		func(settings *gremlin.DriverRemoteConnectionSettings) {
			settings.ConnectionTimeout = cfg.JanusGraph.ConnectionTimeout
			settings.LogVerbosity = gremlin.Warning
		},
	)
	if err != nil {
		return nil, err
	}

	jgp := &JanusGraphProvider{
		cfg:  cfg,
		drc:  driver,
		tags: tag.GetBaseTagsWith(tag.Storage(StorageProviderName)),
	}

	return jgp, nil
}

func (jgp *JanusGraphProvider) Name() string {
	return StorageProviderName
}

func (jgp *JanusGraphProvider) Prepare(ctx context.Context) error {
	if !jgp.cfg.Storage.Wipe {
		log.Trace(ctx).Warn("Skipping graph vertex wipe")

		return nil
	}

	g := gremlin.Traversal_().WithRemote(jgp.drc)
	tx := g.Tx()
	defer tx.Close()

	gtx, err := tx.Begin()
	if err != nil {
		return err
	}

	err = <-gtx.V().Drop().Iterate()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// HealthCheck sends a single digit, as a string. JanusGraph will reply to this with the same value (arithmetic operation)
// We choose the value "1" because it's not the default int value in case there's an issue somewhere.
// from: https://stackoverflow.com/questions/59396980/gremlin-query-to-check-connection-health
func (jgp *JanusGraphProvider) HealthCheck(ctx context.Context) (bool, error) {
	wantValue := "1"
	if jgp.drc == nil {
		return false, errors.New("get janus graph client (nil)")
	}
	res, err := jgp.drc.Submit(wantValue)
	if err != nil {
		return false, err
	}

	one, ok, err := res.One()
	if !ok || err != nil {
		return false, fmt.Errorf("get one results from healthcheck, got: %s", one)
	}

	value, err := one.GetInt()
	if err != nil {
		return false, fmt.Errorf("get int value from healthcheck: %w", err)
	}

	if value != 1 {
		log.Trace(ctx).Errorf("healthcheck returned wrong value, got: %d wanted: %s", value, wantValue)

		return false, nil
	}

	return true, nil
}

// Raw returns a handle to the underlying provider to allow implementation specific operations e.g graph queries.
func (jgp *JanusGraphProvider) Raw() any {
	return jgp.drc
}

// VertexWriter creates a new AsyncVertexWriter instance to enable asynchronous bulk inserts of vertices.
func (jgp *JanusGraphProvider) VertexWriter(ctx context.Context, v vertex.Builder,
	c cache.CacheProvider, opts ...WriterOption) (AsyncVertexWriter, error) {

	opts = append(opts, WithTags(jgp.tags))

	return NewJanusGraphAsyncVertexWriter(ctx, jgp.drc, v, c, opts...)
}

// EdgeWriter creates a new AsyncEdgeWriter instance to enable asynchronous bulk inserts of edges.
func (jgp *JanusGraphProvider) EdgeWriter(ctx context.Context, e edge.Builder, opts ...WriterOption) (AsyncEdgeWriter, error) {
	opts = append(opts, WithTags(jgp.tags))

	return NewJanusGraphAsyncEdgeWriter(ctx, jgp.drc, e, opts...)
}

// Close cleans up any resources used by the Provider implementation. Provider cannot be reused after this call.
func (jgp *JanusGraphProvider) Close(ctx context.Context) error {
	jgp.drc.Close()

	return nil
}

// Raw returns a handle to the underlying provider to allow implementation specific operations e.g graph queries.
func (jgp *JanusGraphProvider) Clean(ctx context.Context, cluster string) error {
	var err error
	span, ctx := span.SpanRunFromContext(ctx, span.IngestorClean)
	defer func() { span.Finish(tracer.WithError(err)) }()
	l := log.Trace(ctx)
	l.Infof("Cleaning cluster", log.FieldClusterKey, cluster)
	g := gremlin.Traversal_().WithRemote(jgp.drc)
	tx := g.Tx()
	defer tx.Close()

	gtx, err := tx.Begin()
	if err != nil {
		return err
	}

	err = <-gtx.V().Has("cluster", cluster).Drop().Iterate()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
