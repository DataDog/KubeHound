package api

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	mocksNotifier "github.com/DataDog/KubeHound/pkg/ingestor/notifier/mocks"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/blob"
	mocksPuller "github.com/DataDog/KubeHound/pkg/ingestor/puller/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/providers"
	mocksCache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	mocksGraph "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	mocksStore "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/stretchr/testify/mock"
)

func TestIngestorAPI_Ingest(t *testing.T) {
	t.Parallel()
	type fields struct {
		cfg *config.KubehoundConfig
	}
	type args struct {
		clusterName string
		runID       string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		mock    func(puller *mocksPuller.DataPuller, notifier *mocksNotifier.Notifier, cache *mocksCache.CacheProvider, store *mocksStore.Provider, graph *mocksGraph.Provider)
	}{
		{
			name: "Pulling invalid bucket name",
			fields: fields{
				cfg: config.MustLoadEmbedConfig(),
			},
			args: args{
				clusterName: "test-cluster",
				runID:       "test-run-id",
			},
			wantErr: true,
			mock: func(puller *mocksPuller.DataPuller, notifier *mocksNotifier.Notifier, cache *mocksCache.CacheProvider, store *mocksStore.Provider, graph *mocksGraph.Provider) {
				puller.On("Pull", mock.Anything, "test-cluster", "test-run-id").Return("", blob.ErrInvalidBucketName)
			},
		},
		// // TODO: find a better way to test this
		// // The mock here would be very fragile and annoying to use: it depends on ~all the mocks of KH.
		// // (we need to mock all the datastore, the writers, the graph builder...)
		// // Keeping this as an example to understand the issues and potentially as an head start for improvement later.
		// {
		// 	name: "Pull archive successfully",
		// 	fields: fields{
		// 		cfg: config.MustLoadEmbedConfig(),
		// 	},
		// 	args: args{
		// 		clusterName: "test-cluster",
		// 		runID:       "test-run-id",
		// 	},
		// 	wantErr: false,
		// 	mock: func(puller *mocksPuller.DataPuller, notifier *mocksNotifier.Notifier, cache *mocksCache.CacheProvider, store *mocksStore.Provider, graph *mocksGraph.Provider, edgeWriter *mocksGraph.AsyncEdgeWriter, cacheReader *mocksCache.CacheReader) {
		// 		puller.EXPECT().Pull(mock.Anything, "test-cluster", "test-run-id").Return("/tmp/kubehound/kh-1234/test-cluster/test-run-id", nil)
		// 		puller.EXPECT().Close(mock.Anything, "/tmp/kubehound/kh-1234/test-cluster/test-run-id").Return(nil)
		// 		puller.EXPECT().Extract(mock.Anything, "/tmp/kubehound/kh-1234/test-cluster/test-run-id").Return(nil)
		// 		store.On("HealthCheck", mock.Anything).Return(true, nil)
		// 		graph.On("HealthCheck", mock.Anything).Return(true, nil)
		// 		cache.On("HealthCheck", mock.Anything).Return(true, nil)
		// 	},
		// },
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockedPuller := mocksPuller.NewDataPuller(t)
			mockedNotifier := mocksNotifier.NewNotifier(t)
			mockedCache := mocksCache.NewCacheProvider(t)
			mockedStoreDB := mocksStore.NewProvider(t)
			mockedGraphDB := mocksGraph.NewProvider(t)

			mockedProvider := &providers.ProvidersFactoryConfig{
				CacheProvider: mockedCache,
				StoreProvider: mockedStoreDB,
				GraphProvider: mockedGraphDB,
			}

			g := NewIngestorAPI(tt.fields.cfg, mockedPuller, mockedNotifier, mockedProvider)
			tt.mock(mockedPuller, mockedNotifier, mockedCache, mockedStoreDB, mockedGraphDB)
			if err := g.Ingest(context.TODO(), tt.args.clusterName, tt.args.runID); (err != nil) != tt.wantErr {
				t.Errorf("IngestorAPI.Ingest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
