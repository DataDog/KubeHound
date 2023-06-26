package graph

// import (
// 	"context"
// 	"errors"
// 	"strconv"
// 	"testing"

// 	e "github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
// 	edge "github.com/DataDog/KubeHound/pkg/kubehound/graph/edge/mocks"
// 	cache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
// 	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
// 	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// func NewTestBuilder(t *testing.T, mockRegistry e.EdgeRegistry, gdb *graphdb.Provider, sdb *storedb.Provider) *Builder {
// 	builder := &Builder{
// 		storedb: sdb,
// 		graphdb: gdb,
// 		edges:   mockRegistry,
// 	}

// 	return builder
// }

// func TestGraphBuilder_Success(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()
// 	gdb := graphdb.NewProvider(t)
// 	sdb := storedb.NewProvider(t)
// 	c := cache.NewCacheReader(t)

// 	const numTestEdges = 30
// 	reg := make(e.EdgeRegistry)
// 	for i := 0; i < numTestEdges; i++ {
// 		e := edge.NewBuilder(t)

// 		e.EXPECT().Stream(mock.Anything, sdb, c, mock.AnythingOfType("edge.ProcessEntryCallback"), mock.AnythingOfType("edge.CompleteQueryCallback")).Return(nil)
// 		reg["EDGE_"+strconv.Itoa(i)] = e
// 	}

// 	b := NewTestBuilder(t, reg, gdb, sdb)
// 	gw := graphdb.NewAsyncEdgeWriter(t)
// 	gw.EXPECT().Close(mock.Anything).Return(nil).Times(numTestEdges)

// 	for _, ee := range reg {
// 		e := ee
// 		gdb.EXPECT().EdgeWriter(mock.Anything, e).Return(gw, nil)
// 	}

// 	err := b.Run(ctx)
// 	assert.NoError(t, err)
// }

// func TestGraphBuilder_HealthCheck(t *testing.T) {
// 	t.Parallel()

// 	gdb := graphdb.NewProvider(t)
// 	sdb := storedb.NewProvider(t)
// 	reg := make(e.EdgeRegistry)

// 	b := NewTestBuilder(t, reg, gdb, sdb)

// 	// All ok
// 	gdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()
// 	sdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()

// 	assert.NoError(t, b.HealthCheck(context.TODO()))

// 	// Graph failure
// 	gdb.EXPECT().Name().Return("graph")
// 	gdb.EXPECT().HealthCheck(mock.Anything).Return(false, errors.New("service error")).Once()
// 	sdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()

// 	assert.ErrorContains(t, b.HealthCheck(context.TODO()), "service error")

// 	// Store failure
// 	sdb.EXPECT().Name().Return("store")
// 	gdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()
// 	sdb.EXPECT().HealthCheck(mock.Anything).Return(false, nil).Once()

// 	assert.ErrorContains(t, b.HealthCheck(context.TODO()), "store healthcheck")

// 	// Friday 5pm :)
// 	gdb.EXPECT().Name().Return("graph")
// 	sdb.EXPECT().Name().Return("store")
// 	gdb.EXPECT().HealthCheck(mock.Anything).Return(false, nil).Once()
// 	sdb.EXPECT().HealthCheck(mock.Anything).Return(false, nil).Once()

// 	assert.ErrorContains(t, b.HealthCheck(context.TODO()), "2 errors occurred:")
// }

// func TestGraphBuilder_EdgeErrorCancelsAll(t *testing.T) {
// 	t.Parallel()

// 	gdb := graphdb.NewProvider(t)
// 	sdb := storedb.NewProvider(t)
// 	c := cache.NewCacheReader(t)

// 	const numTestEdges = 30
// 	reg := make(e.EdgeRegistry)
// 	for i := 0; i < numTestEdges; i++ {
// 		e := edge.NewBuilder(t)

// 		switch {
// 		case i == 5:
// 			// Raise an errors. Must be called
// 			e.EXPECT().Stream(mock.Anything, sdb, c, mock.AnythingOfType("edge.ProcessEntryCallback"), mock.AnythingOfType("edge.CompleteQueryCallback")).Return(errors.New("test error"))
// 		default:
// 			// No errors. May or May not be called!
// 			e.EXPECT().Stream(mock.Anything, sdb, c, mock.AnythingOfType("edge.ProcessEntryCallback"), mock.AnythingOfType("edge.CompleteQueryCallback")).Return(nil).Maybe()
// 		}

// 		reg["EDGE_"+strconv.Itoa(i)] = e
// 	}

// 	b := NewTestBuilder(t, reg, gdb, sdb)
// 	gw := graphdb.NewAsyncEdgeWriter(t)
// 	gdb.EXPECT().EdgeWriter(mock.Anything, mock.Anything).Return(gw, nil).Maybe()
// 	gw.EXPECT().Close(mock.Anything).Return(nil).Maybe()

// 	err := b.Run(context.Background())
// 	assert.ErrorContains(t, err, "test error")
// }
