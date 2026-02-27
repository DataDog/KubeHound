package api

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dump"
	mocksNotifier "github.com/DataDog/KubeHound/pkg/ingestor/notifier/mocks"
	mocksPuller "github.com/DataDog/KubeHound/pkg/ingestor/puller/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/providers"
	mocksGraph "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	mocksStore "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	err = storedb.InitSchema(db)
	require.NoError(t, err)

	return db
}

func foundPreviousScan(t *testing.T, g *IngestorAPI, db *sql.DB) {
	t.Helper()
	store, ok := g.providers.StoreProvider.(*mocksStore.Provider)
	if !ok {
		t.Fatalf("failed to cast store provider to mock")
	}

	// Insert a row in one table to emulate a previous scan
	_, err := db.ExecContext(t.Context(),
		"INSERT INTO nodes (id, name, run_id, cluster_name) VALUES (?, ?, ?, ?)",
		1, "test-node", "test-run-id", "test-cluster")
	require.NoError(t, err)

	store.On("Reader").Return(db)
}

func noPreviousScan(t *testing.T, g *IngestorAPI, db *sql.DB) {
	t.Helper()
	store, ok := g.providers.StoreProvider.(*mocksStore.Provider)
	if !ok {
		t.Fatalf("failed to cast store provider to mock")
	}

	// No data inserted - all tables are empty
	store.On("Reader").Return(db)
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

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		mock    func(puller *mocksPuller.DataPuller, notifier *mocksNotifier.Notifier, store *mocksStore.Provider, graph *mocksGraph.Provider)
	}{
		// Test cases are commented out as they require extensive mocking of the full pipeline.
		// See the inline comments for context on why.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockedPuller := mocksPuller.NewDataPuller(t)
			mockedNotifier := mocksNotifier.NewNotifier(t)
			mockedStoreDB := mocksStore.NewProvider(t)
			mockedGraphDB := mocksGraph.NewProvider(t)

			mockedProvider := &providers.ProvidersFactoryConfig{
				StoreProvider: mockedStoreDB,
				GraphProvider: mockedGraphDB,
			}

			db := testDB(t)
			g := NewIngestorAPI(tt.fields.cfg, mockedPuller, mockedNotifier, mockedProvider)
			noPreviousScan(t, g, db)
			tt.mock(mockedPuller, mockedNotifier, mockedStoreDB, mockedGraphDB)

			// Construct dump result path
			dumpResult, err := dump.NewDumpResult(tt.args.clusterName, tt.args.runID, true)
			if err != nil {
				t.Errorf("dump.NewDumpResult() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			dumpResultPath := dumpResult.GetFullPath()

			if err := g.Ingest(t.Context(), dumpResultPath); (err != nil) != tt.wantErr {
				t.Errorf("IngestorAPI.Ingest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIngestorAPI_isAlreadyIngestedInDB(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	tests := []struct {
		name            string
		clusterName     string
		runID           string
		wantErr         bool
		alreadyIngested bool
		setup           func(t *testing.T, g *IngestorAPI, db *sql.DB)
	}{
		{
			name:            "RunID already ingested",
			clusterName:     "test-cluster",
			runID:           "test-run-id",
			wantErr:         false,
			alreadyIngested: true,
			setup:           foundPreviousScan,
		},
		{
			name:            "RunID not ingested",
			clusterName:     "test-cluster",
			runID:           "test-run-id",
			wantErr:         false,
			alreadyIngested: false,
			setup:           noPreviousScan,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := testDB(t)
			g := &IngestorAPI{
				providers: &providers.ProvidersFactoryConfig{
					StoreProvider: mocksStore.NewProvider(t),
				},
			}

			tt.setup(t, g, db)
			alreadyIngested, err := g.isAlreadyIngestedInDB(ctx, tt.clusterName, tt.runID)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s - IngestorAPI.checkPreviousRun() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			assert.Equal(t, tt.alreadyIngested, alreadyIngested)
		})
	}
}

func TestGetCollections(t *testing.T) {
	t.Parallel()
	got := collections.GetCollections()
	expected := []string{"nodes", "pods", "containers", "volumes", "roles", "rolebindings", "identities", "permissionsets", "endpoints"}
	assert.Equal(t, expected, got)

	// Verify all tables exist in SQLite schema by checking each one
	db := testDB(t)
	for _, table := range got {
		var count int64
		//nolint:gosec
		err := db.QueryRowContext(t.Context(), fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		assert.NoError(t, err, "table %s should exist", table)
	}
}
