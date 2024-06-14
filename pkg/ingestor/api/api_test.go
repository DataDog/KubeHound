package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	mocksNotifier "github.com/DataDog/KubeHound/pkg/ingestor/notifier/mocks"
	mocksPuller "github.com/DataDog/KubeHound/pkg/ingestor/puller/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/providers"
	mocksCache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	mocksGraph "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	mocksStore "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func foundPreviousScan(mt *mtest.T, g *IngestorAPI) {
	mt.Helper()

	store, ok := g.providers.StoreProvider.(*mocksStore.Provider)
	if !ok {
		mt.Fatalf("failed to cast store provider to mock")
	}

	defer func() {
		mt.Client = nil
	}()
	mt.AddMockResponses(mtest.CreateSuccessResponse())

	// Return X documents to emulate a previous scan
	store.On("Reader", mock.Anything).Return(mt.DB).Once()
	mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.nodes", mtest.FirstBatch, bson.D{{Key: "n", Value: 123}}))

	// For count to work, mongo needs an index. So we need to create that. Index view should contains a key. Value does not matter
	indexView := mt.Coll.Indexes()
	_, err := indexView.CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{{Key: "n", Value: 321}},
	})
	require.Nil(mt, err, "CreateOne error for index: %v", err)
}

func noPreviousScan(mt *mtest.T, g *IngestorAPI) {
	mt.Helper()

	store, ok := g.providers.StoreProvider.(*mocksStore.Provider)
	if !ok {
		mt.Fatalf("failed to cast store provider to mock")
	}

	defer func() {
		mt.Client = nil
	}()
	mt.AddMockResponses(mtest.CreateSuccessResponse())

	// For count to work, mongo needs an index. So we need to create that. Index view should contains a key. Value does not matter
	indexView := mt.Coll.Indexes()
	_, err := indexView.CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{{Key: "n", Value: 321}},
	})
	require.Nil(mt, err, "CreateOne error for index: %v", err)

	// Iterate over all collections without findings any element
	for _, collection := range collections.GetCollections() {
		store.On("Reader", mock.Anything).Return(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("test.%s", collection), mtest.FirstBatch, bson.D{{Key: "n", Value: 0}}))

	}
}

func TestIngestorAPI_Ingest(t *testing.T) {
	t.Parallel()
	type fields struct {
		cfg *config.KubehoundConfig
	}
	type args struct {
		clusterName string
		runID       string
	}

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		mock    func(puller *mocksPuller.DataPuller, notifier *mocksNotifier.Notifier, cache *mocksCache.CacheProvider, store *mocksStore.Provider, graph *mocksGraph.Provider)
	}{
		// {
		// 	name: "Pulling invalid bucket name",
		// 	fields: fields{
		// 		cfg: config.MustLoadEmbedConfig(),
		// 	},
		// 	args: args{
		// 		clusterName: "test-cluster",
		// 		runID:       "test-run-id",
		// 	},
		// 	wantErr: true,
		// 	mock: func(puller *mocksPuller.DataPuller, notifier *mocksNotifier.Notifier, cache *mocksCache.CacheProvider, store *mocksStore.Provider, graph *mocksGraph.Provider) {
		// 		puller.On("Pull", mock.Anything, "test-cluster", "test-run-id").Return("", blob.ErrInvalidBucketName)
		// 	},
		// },
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
		mt.Run(tt.name, func(mt *mtest.T) {
			mt.Parallel()
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
			noPreviousScan(mt, g)
			tt.mock(mockedPuller, mockedNotifier, mockedCache, mockedStoreDB, mockedGraphDB)
			if err := g.Ingest(context.TODO(), tt.args.clusterName, tt.args.runID); (err != nil) != tt.wantErr {
				t.Errorf("IngestorAPI.Ingest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIngestorAPI_isAlreadyIngestedInDB(t *testing.T) {
	t.Parallel()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	ctx := context.TODO()

	type fields struct {
		providers *providers.ProvidersFactoryConfig
		mock      *mtest.T
	}
	type args struct {
		clusterName string
		runID       string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantErr         bool
		alreadyIngested bool
		testfct         func(mt *mtest.T, g *IngestorAPI)
	}{
		{
			name: "RunID already ingested",
			fields: fields{
				mock: mt,
				providers: &providers.ProvidersFactoryConfig{
					StoreProvider: mocksStore.NewProvider(t),
				},
			},
			args: args{
				clusterName: "test-cluster",
				runID:       "test-run-id",
			},
			wantErr:         false,
			testfct:         foundPreviousScan,
			alreadyIngested: true,
		},
		{
			name: "RunID not ingested",
			fields: fields{
				mock: mt,
				providers: &providers.ProvidersFactoryConfig{
					StoreProvider: mocksStore.NewProvider(t),
				},
			},
			args: args{
				clusterName: "test-cluster",
				runID:       "test-run-id",
			},
			wantErr:         false,
			testfct:         noPreviousScan,
			alreadyIngested: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		mt.Run(tt.name, func(mt *mtest.T) {
			mt.Parallel()
			g := &IngestorAPI{
				providers: tt.fields.providers,
			}

			tt.testfct(mt, g)
			alreadyIngested, err := g.isAlreadyIngestedInDB(ctx, tt.args.clusterName, tt.args.runID)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s - IngestorAPI.checkPreviousRun() error = %d, wantErr %v", tt.name, err, tt.wantErr)
			}
			assert.Equal(t, tt.alreadyIngested, alreadyIngested)
		})
	}
}
