package ingestor

import (
	"context"
	"errors"
	"testing"

	collector "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/pipeline"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/pipeline/mocks"
	cache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPipelineIngestor_HealthCheck(t *testing.T) {
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

func createMockData(t *testing.T) (map[string]*mocks.ObjectIngest, *PipelineIngestor) {
	ingests := make(map[string]*mocks.ObjectIngest)
	ingests["nodes"] = mocks.NewObjectIngest(t)
	ingests["pods"] = mocks.NewObjectIngest(t)
	ingests["roles"] = mocks.NewObjectIngest(t)
	ingests["croles"] = mocks.NewObjectIngest(t)
	ingests["rolebindings"] = mocks.NewObjectIngest(t)
	ingests["crolebindings"] = mocks.NewObjectIngest(t)

	seq := []pipeline.Sequence{
		{
			Name: "identity-pipeline",
			Groups: []pipeline.Group{
				{
					Name: "k8s-role-group",
					Ingests: []pipeline.ObjectIngest{
						ingests["roles"],
						ingests["croles"],
					},
				},
				{
					Name: "k8s-rolebinding-group",
					Ingests: []pipeline.ObjectIngest{
						ingests["rolebindings"],
						ingests["crolebindings"],
					},
				},
			},
		},
		{
			Name: "application-pipeline",
			Groups: []pipeline.Group{
				{
					Name: "k8s-node-group",
					Ingests: []pipeline.ObjectIngest{
						ingests["nodes"],
					},
				},
				{
					Name: "k8s-pod-group",
					Ingests: []pipeline.ObjectIngest{
						ingests["pods"],
					},
				},
			},
		},
	}

	pi := &PipelineIngestor{
		sequences: seq,
	}

	return ingests, pi
}

func TestPipelineIngestor_Run(t *testing.T) {
	t.Parallel()

	ingests, pi := createMockData(t)

	for key, ingest := range ingests {
		oi := ingest
		oi.EXPECT().Name().Return(key)
		oi.EXPECT().Initialize(mock.Anything, mock.AnythingOfType("*pipeline.Dependencies")).Return(nil).Once()
		oi.EXPECT().Close(mock.Anything).Return(nil).Once()
	}

	// Expect calls to each object respecting the order/parallel logic of the pipeline
	nodeRun := ingests["nodes"].EXPECT().Run(mock.Anything).Return(nil).Once()
	ingests["pods"].EXPECT().Run(mock.Anything).Return(nil).Once().NotBefore(nodeRun)

	roleRun := ingests["roles"].EXPECT().Run(mock.Anything).Return(nil).Once()
	croleRun := ingests["croles"].EXPECT().Run(mock.Anything).Return(nil).Once()
	ingests["rolebindings"].EXPECT().Run(mock.Anything).Return(nil).Once().NotBefore(roleRun).NotBefore(croleRun)
	ingests["crolebindings"].EXPECT().Run(mock.Anything).Return(nil).Once().NotBefore(roleRun).NotBefore(croleRun)

	err := pi.Run(context.Background())
	assert.NoError(t, err)
}

func TestPipelineIngestor_RunInitError(t *testing.T) {
	t.Parallel()

	ingests, pi := createMockData(t)
	for key, ingest := range ingests {
		oi := ingest
		switch key {
		case "nodes":
			oi.EXPECT().Initialize(mock.Anything, mock.AnythingOfType("*pipeline.Dependencies")).Return(errors.New("test error")).Once()
			oi.EXPECT().Name().Return(key)
		case "pods":
			// None - node ingest error means pods should be called!
		default:
			// May or may not be called
			oi.EXPECT().Initialize(mock.Anything, mock.AnythingOfType("*pipeline.Dependencies")).Return(nil).Maybe()
			oi.EXPECT().Name().Return(key).Maybe()
			oi.EXPECT().Close(mock.Anything).Return(nil).Maybe()
		}
	}

	// Expect calls to each object respecting the order/parallel logic of the pipeline
	roleRun := ingests["roles"].EXPECT().Run(mock.Anything).Return(nil).Maybe()
	croleRun := ingests["croles"].EXPECT().Run(mock.Anything).Return(nil).Maybe()
	ingests["rolebindings"].EXPECT().Run(mock.Anything).Maybe().NotBefore(roleRun).NotBefore(croleRun)
	ingests["crolebindings"].EXPECT().Run(mock.Anything).Return(nil).Maybe().NotBefore(roleRun).NotBefore(croleRun)

	err := pi.Run(context.Background())
	assert.ErrorContains(t, err, "group k8s-node-group ingest: test error")
}

func TestPipelineIngestor_RunExecError(t *testing.T) {
	t.Parallel()

	ingests, pi := createMockData(t)
	for key, ingest := range ingests {
		oi := ingest

		switch key {
		case "rolebindings", "crolebindings":
			// Should not be initialize (or run) due to run failure of previous stage
		default:
			// May or may not be called
			oi.EXPECT().Initialize(mock.Anything, mock.AnythingOfType("*pipeline.Dependencies")).Return(nil).Maybe()
			oi.EXPECT().Name().Return(key).Maybe()
			oi.EXPECT().Close(mock.Anything).Return(nil).Maybe()
		}
	}

	// Expect calls to each object respecting the order/parallel logic of the pipeline
	nodeRun := ingests["nodes"].EXPECT().Run(mock.Anything).Return(nil).Maybe()
	ingests["pods"].EXPECT().Run(mock.Anything).Return(nil).Maybe().NotBefore(nodeRun)

	ingests["roles"].EXPECT().Run(mock.Anything).Return(nil).Once()
	ingests["croles"].EXPECT().Run(mock.Anything).Return(errors.New("test error")).Once()

	err := pi.Run(context.Background())
	assert.ErrorContains(t, err, "group k8s-role-group ingest: test error")
}
