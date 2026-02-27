//nolint:forcetypeassert
package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollect "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClusterRoleIngest_Pipeline(t *testing.T) {
	t.Parallel()

	cri := &ClusterRoleIngest{}

	ctx := t.Context()
	fakeRole, err := loadTestObject[types.ClusterRoleType]("testdata/clusterrole.json")
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamClusterRoles(ctx, cri).
		RunAndReturn(func(ctx context.Context, i collector.ClusterRoleIngestor) error {
			// Fake the stream of a single cluster role from the collector client
			err := i.IngestClusterRole(ctx, fakeRole)
			if err != nil {
				return err
			}

			return i.Complete(ctx)
		})

	// Store setup
	sdb := storedb.NewProvider(t)
	storeID := store.ObjectID()
	sdb.EXPECT().Write(ctx, mock.AnythingOfType("*store.Role")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Role).Id = storeID

			return nil
		}).Once()

	deps := &Dependencies{
		Collector: client,
		StoreDB:   sdb,
		Config: &config.KubehoundConfig{
			Builder: config.BuilderConfig{
				Edge: config.EdgeBuilderConfig{},
			},
			Dynamic: config.DynamicConfig{
				RunID: config.NewRunID(),
				Cluster: config.DynamicClusterInfo{
					Name: "test-cluster",
				},
			},
		},
	}

	// Initialize
	err = cri.Initialize(ctx, deps)
	assert.NoError(t, err)

	// Run
	err = cri.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = cri.Close(ctx)
	assert.NoError(t, err)
}
