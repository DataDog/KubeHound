package edge

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func init() {
	Register(&EscapeVarLogSymlink{}, RegisterGraphDependency)
}

type EscapeVarLogSymlink struct {
	BaseContainerEscape
}

// The query returns a list of permissionSet IDs
type permissionSetIDEscapeGroup struct {
	PermissionSetID int64 `json:"permission_set"`
}

func (e *EscapeVarLogSymlink) Label() string {
	return "CE_VAR_LOG_SYMLINK"
}

// List of needed edges to run the traversal query
func (e *EscapeVarLogSymlink) Dependencies() []string {
	return []string{"PERMISSION_DISCOVER", "IDENTITY_ASSUME", "VOLUME_DISCOVER", "VOLUME_ACCESS"}
}

func (e *EscapeVarLogSymlink) Name() string {
	return "ContainerEscapeVarLogSymlink"
}

func (e *EscapeVarLogSymlink) AttckTechniqueID() AttckTechniqueID {
	return AttckTechniqueUnsecuredCredentials
}

func (e *EscapeVarLogSymlink) AttckTacticID() AttckTacticID {
	return AttckTacticCredentialAccess
}

func (e *EscapeVarLogSymlink) Processor(ctx context.Context, oic *converter.ObjectIDConverter, entry any) (any, error) {
	typed, ok := entry.(*permissionSetIDEscapeGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	permissionSetVertexID, err := oic.GraphID(ctx, store.Hex(typed.PermissionSetID))
	if err != nil {
		return nil, fmt.Errorf("%s edge IN id convert: %w", e.Label(), err)
	}

	return permissionSetVertexID, nil
}

func (e *EscapeVarLogSymlink) Traversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []any) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()
		// reduce the graph to only these permission sets
		g.V(inserts...).Has("class", "PermissionSet").
			// get identity vertices
			InE("PERMISSION_DISCOVER").OutV().
			// get container vertices
			InE("IDENTITY_ASSUME").OutV().
			// save container vertices as "c" so we can link to it to the node via CE_VAR_LOG_SYMLINK
			Has("class", "Container").As("c").
			// Get all the volumes
			OutE("VOLUME_DISCOVER").InV().
			Has("type", "HostPath").
			// filter only the volumes that are "affected" by this attacks ("/", "/var", "/var/log").
			Has("sourcePath", P.Within("/", "/var", "/var/log")).
			// get the node related to that volume mount
			InE("VOLUME_ACCESS").OutV().
			Has("class", "Node").As("n").
			AddE("CE_VAR_LOG_SYMLINK").From("c").To("n").
			Property("attckTechniqueID", string(e.AttckTechniqueID())).
			Property("attckTacticID", string(e.AttckTacticID())).
			Barrier().Limit(0)

		return g
	}
}

func (e *EscapeVarLogSymlink) Stream(ctx context.Context, _ storedb.Provider, db *sql.DB,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback,
) error {

	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT ps.id
		FROM permissionsets ps, json_each(ps.rules) AS r
		WHERE ps.run_id = ? AND ps.cluster_name = ?
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.apiGroups')) WHERE value IN ('', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.resources')) WHERE value IN ('pods/log', 'pods/*', '*'))
		AND EXISTS (SELECT 1 FROM json_each(json_extract(r.value, '$.verbs')) WHERE value IN ('get', '*'))
		AND (json_extract(r.value, '$.resourceNames') IS NULL OR json_extract(r.value, '$.resourceNames') = '[]')`,
		e.runtime.RunID.String(), e.runtime.Cluster.Name)
	if err != nil {
		return err
	}

	return adapter.SQLiteRowHandler[permissionSetIDEscapeGroup](ctx, rows, func(row *sql.Rows) (permissionSetIDEscapeGroup, error) {
		var g permissionSetIDEscapeGroup
		err := row.Scan(&g.PermissionSetID)
		return g, err
	}, callback, complete)
}
