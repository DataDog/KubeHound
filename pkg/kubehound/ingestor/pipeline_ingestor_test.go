package ingestor

import (
	"context"
	"errors"
	"testing"

	collector "github.com/DataDog/KubeHound/pkg/collector/mocks"
	"github.com/DataDog/KubeHound/pkg/config"
	cache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Ensure on happy path everything is called in expected order
// Ensure on error path everything is cancelled appropriately
func TestPipelineIngestor_HealtCheck(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := collector.NewCollectorClient(t)
	c := cache.NewCacheProvider(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	i, err := newPipelineIngestor(&config.KubehoundConfig{}, client, c, sdb, gdb)
	assert.NoError(t, err)

	// All succeeded
	client.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()
	c.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()
	gdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()
	sdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()

	assert.NoError(t, i.HealthCheck(ctx))

	// Multiple failures
	client.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()
	c.EXPECT().HealthCheck(mock.Anything).Return(false, nil).Once()
	c.EXPECT().Name().Return("cache").Once()
	gdb.EXPECT().HealthCheck(mock.Anything).Return(true, nil).Once()
	sdb.EXPECT().HealthCheck(mock.Anything).Return(false, errors.New("store test")).Once()

	err = i.HealthCheck(ctx)
	assert.ErrorContains(t, err, "cache healthcheck")
	assert.ErrorContains(t, err, "store test")
}

func TestPipelineIngestor_Run(t *testing.T) {
	t.Parallel()

}

func TestPipelineIngestor_RunError(t *testing.T) {
	t.Parallel()

}
