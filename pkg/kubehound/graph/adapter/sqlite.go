package adapter

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
)

// SQLiteRowHandler is the default stream implementation to handle query results from a SQLite database.
// It iterates over rows, scanning each into a T via the provided scan function,
// and passes each entry to the callback. When iteration is done (or on error), it calls complete.
func SQLiteRowHandler[T any](ctx context.Context, rows *sql.Rows, scan func(*sql.Rows) (T, error),
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	defer rows.Close()

	var lastErr error
	for rows.Next() {
		entry, err := scan(rows)
		if err != nil {
			lastErr = err
			break
		}

		lastErr = callback(ctx, &entry)
		if lastErr != nil {
			break
		}
	}

	err := complete(ctx)
	return errors.Join(err, lastErr)
}
