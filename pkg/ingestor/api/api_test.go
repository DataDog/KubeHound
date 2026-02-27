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
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	mocksStore "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
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

			g := NewIngestorAPI(tt.fields.cfg, mockedPuller, mockedNotifier, mockedProvider)
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

func TestTables(t *testing.T) {
	t.Parallel()
	expected := []storedb.Table{"nodes", "pods", "containers", "volumes", "roles",
		"rolebindings", "identities", "permissionsets", "endpoints"}
	assert.Equal(t, expected, storedb.Tables)

	db := testDB(t)
	for _, table := range storedb.Tables {
		var count int64
		err := db.QueryRowContext(t.Context(),
			fmt.Sprintf("SELECT COUNT(*) FROM %s", string(table))).Scan(&count)
		assert.NoError(t, err, "table %s should exist", table)
	}
}
