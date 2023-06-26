package graphdb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/path"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

// TODO maybe move that into a config file?
const channelSizeBatchFactor = 4

var _ Provider = (*JanusGraphProvider)(nil)

type JanusGraphProvider struct {
	pool *DriverConnectionPool
	tags []string
}

func NewGraphDriver(ctx context.Context, dbHost string, timeout time.Duration) (*JanusGraphProvider, error) {
	if dbHost == "" {
		return nil, errors.New("JanusGraph DB URL is not set")
	}
	driver, err := gremlingo.NewDriverRemoteConnection(dbHost,
		func(settings *gremlingo.DriverRemoteConnectionSettings) {
			settings.ConnectionTimeout = timeout
		},
	)
	if err != nil {
		return nil, err
	}

	g := &JanusGraphProvider{
		pool: &DriverConnectionPool{
			Lock:   &sync.Mutex{},
			Driver: driver,
		},
		tags: append(telemetry.BaseTags, telemetry.TagTypeJanusGraph),
	}

	return g, nil
}

func (jgp *JanusGraphProvider) Name() string {
	return "JanusGraphProvider"
}

// HealthCheck sends a single digit, as a string. JanusGraph will reply to this with the same value (arithmetic operation)
// We choose the value "1" because it's not the default int value in case there's an issue somewhere.
// from: https://stackoverflow.com/questions/59396980/gremlin-query-to-check-connection-health
func (jgp *JanusGraphProvider) HealthCheck(ctx context.Context) (bool, error) {
	wantValue := "1"
	if jgp.pool.Driver == nil {
		return false, errors.New("get janus graph client (nil)")
	}
	res, err := jgp.pool.Driver.Submit(wantValue)
	if err != nil {
		return false, err
	}

	one, ok, err := res.One()
	if !ok || err != nil {
		return false, fmt.Errorf("get one results from healthcheck, got: %s", one)
	}

	value, err := one.GetInt()
	if err != nil {
		return false, fmt.Errorf("get int value from healthcheck: %v", err)
	}

	if value != 1 {
		log.I.Errorf("healthcheck returned wrong value, got: %d wanted: %s", value, wantValue)
		return false, nil
	}

	return true, nil
}

// Raw returns a handle to the underlying provider to allow implementation specific operations e.g graph queries.
func (jgp *JanusGraphProvider) Raw() any {
	return jgp.pool.Driver
}

// VertexWriter creates a new AsyncVertexWriter instance to enable asynchronous bulk inserts of vertices.
func (jgp *JanusGraphProvider) VertexWriter(ctx context.Context, v vertex.Builder, opts ...WriterOption) (AsyncVertexWriter, error) {
	return NewJanusGraphAsyncVertexWriter(ctx, jgp.pool, v, opts...)
}

// EdgeWriter creates a new AsyncEdgeWriter instance to enable asynchronous bulk inserts of edges.
func (jgp *JanusGraphProvider) EdgeWriter(ctx context.Context, e edge.Builder, opts ...WriterOption) (AsyncEdgeWriter, error) {
	return NewJanusGraphAsyncEdgeWriter(ctx, jgp.pool, e, opts...)
}

// PathWriter creates a new AsyncPathWriter instance to enable asynchronous bulk inserts of paths.
func (jgp *JanusGraphProvider) PathWriter(ctx context.Context, p path.Builder, opts ...WriterOption) (AsyncPathWriter, error) {
	return NewJanusGraphAsyncPathWriter(ctx, jgp.pool, p, opts...)
}

// Close cleans up any resources used by the Provider implementation. Provider cannot be reused after this call.
func (jgp *JanusGraphProvider) Close(ctx context.Context) error {
	jgp.pool.Driver.Close()
	return nil
}
