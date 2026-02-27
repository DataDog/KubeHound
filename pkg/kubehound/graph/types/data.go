package types

import (
	"context"
)

// EdgeWriter is an interface for queuing and flushing edge writes.
// This mirrors graphdb.AsyncEdgeWriter but is defined here to avoid import cycles
// between the edge and graphdb packages.
type EdgeWriter interface {
	Queue(ctx context.Context, e any) error
	Flush(ctx context.Context) error
}
