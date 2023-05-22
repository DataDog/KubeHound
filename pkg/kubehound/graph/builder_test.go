package graph

import (
	"context"
	"errors"
	"strconv"
	"testing"

	e "github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	edge "github.com/DataDog/KubeHound/pkg/kubehound/graph/edge/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

func NewTestBuilder(t *testing.T, mockRegistry e.EdgeRegistry, gdb *graphdb.Provider, sdb *storedb.Provider) *Builder {
	builder := &Builder{
		storedb:  sdb,
		graphdb:  gdb,
		registry: mockRegistry,
	}

	return builder
}

func TestGraphBuilder_Success(t *testing.T) {
	t.Parallel()

	fakeTraversal := func(g *gremlin.GraphTraversalSource, inserts []e.TraversalInput) *gremlin.GraphTraversal {
		return nil
	}
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	const numTestEdges = 30
	reg := make(e.EdgeRegistry)
	for i := 0; i < numTestEdges; i++ {
		e := edge.NewEdge(t)

		e.EXPECT().Traversal().Return(fakeTraversal)
		e.EXPECT().Stream(mock.Anything, sdb, mock.AnythingOfType("edge.ProcessEntryCallback"), mock.AnythingOfType("edge.CompleteQueryCallback")).Return(nil)

		reg["EDGE_"+strconv.Itoa(i)] = e
	}

	b := NewTestBuilder(t, reg, gdb, sdb)
	gw := graphdb.NewAsyncEdgeWriter(t)
	gdb.EXPECT().EdgeWriter(mock.Anything, mock.AnythingOfType("edge.EdgeTraversal")).Return(gw, nil)
	gw.EXPECT().Close(mock.Anything).Times(numTestEdges).Return(nil)

	err := b.Run(context.Background())
	assert.NoError(t, err)
}

func TestGraphBuilder_HealthCheck(t *testing.T) {
	t.Parallel()

	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)
	reg := make(e.EdgeRegistry)

	b := NewTestBuilder(t, reg, gdb, sdb)

	// All ok
	gdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()
	sdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()

	assert.NoError(t, b.HealthCheck(context.TODO()))

	// Graph failure
	gdb.EXPECT().Name().Return("graph")
	gdb.EXPECT().HealthCheck(mock.Anything).Return(false, errors.New("service error")).Once()
	sdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()

	assert.ErrorContains(t, b.HealthCheck(context.TODO()), "service error")

	// Store failure
	sdb.EXPECT().Name().Return("store")
	gdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()
	sdb.EXPECT().HealthCheck(mock.Anything).Return(false, nil).Once()

	assert.ErrorContains(t, b.HealthCheck(context.TODO()), "store healthcheck")

	// Friday 5pm :)
	gdb.EXPECT().Name().Return("graph")
	sdb.EXPECT().Name().Return("store")
	gdb.EXPECT().HealthCheck(mock.Anything).Return(false, nil).Once()
	sdb.EXPECT().HealthCheck(mock.Anything).Return(false, nil).Once()

	assert.ErrorContains(t, b.HealthCheck(context.TODO()), "2 errors occurred:")
}

func TestGraphBuilder_EdgeErrorCancelsAll(t *testing.T) {
	t.Parallel()

	fakeTraversal := func(g *gremlin.GraphTraversalSource, inserts []e.TraversalInput) *gremlin.GraphTraversal {
		return nil
	}
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	const numTestEdges = 30
	reg := make(e.EdgeRegistry)
	for i := 0; i < numTestEdges; i++ {
		e := edge.NewEdge(t)

		switch {
		case i == 5:
			// Raise an errors. Must be called
			e.EXPECT().Traversal().Return(fakeTraversal)
			e.EXPECT().Stream(mock.Anything, sdb, mock.AnythingOfType("edge.ProcessEntryCallback"), mock.AnythingOfType("edge.CompleteQueryCallback")).Return(errors.New("test error"))
		default:
			// No errors. May or May not be called!
			e.EXPECT().Traversal().Return(fakeTraversal).Maybe()
			e.EXPECT().Stream(mock.Anything, sdb, mock.AnythingOfType("edge.ProcessEntryCallback"), mock.AnythingOfType("edge.CompleteQueryCallback")).Return(nil).Maybe()
		}

		reg["EDGE_"+strconv.Itoa(i)] = e
	}

	b := NewTestBuilder(t, reg, gdb, sdb)
	gw := graphdb.NewAsyncEdgeWriter(t)
	gdb.EXPECT().EdgeWriter(mock.Anything, mock.AnythingOfType("edge.EdgeTraversal")).Return(gw, nil).Maybe()
	gw.EXPECT().Close(mock.Anything).Return(nil).Maybe()

	err := b.Run(context.Background())
	assert.ErrorContains(t, err, "test error")
}
