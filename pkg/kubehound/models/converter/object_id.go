package converter

import (
	"context"
	"database/sql"
	"fmt"
)

// ObjectIDConverter enables converting between a store object ID and an existing graph vertex ID.
type ObjectIDConverter struct {
	db *sql.DB
}

// NewObjectID creates a new ObjectIdConverter instance from the provided database handle.
func NewObjectID(db *sql.DB) *ObjectIDConverter {
	return &ObjectIDConverter{
		db: db,
	}
}

// GraphID will return the graph vertex ID corresponding to the provided store ID.
func (c *ObjectIDConverter) GraphID(ctx context.Context, storeID string) (int64, error) {
	if c.db == nil {
		return -1, ErrNoDBInitialized
	}

	var vid int64
	err := c.db.QueryRowContext(ctx,
		"SELECT vertex_id FROM store_graph_id_map WHERE store_id = ?", storeID).Scan(&vid)
	if err != nil {
		return -1, fmt.Errorf("graph id lookup (storeID=%s): %w", storeID, err)
	}

	return vid, nil
}
