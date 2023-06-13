package types

import (
	"context"
)

// An object to be consumed by a vertex traversal function to insert a vertex into the graph database.
type TraversalInput any

// An object to encapsulate the raw data required to create one or more edges. For example a pod id and a node id.
type DataContainer any

// ProcessEntryCallback is a callback provided by the the edge builder that will convert edge query results into graph database writes.
type ProcessEntryCallback func(ctx context.Context, model DataContainer) error

// CompleteQueryCallback is a callback provided by the the edge builder that will flush any outstanding graph database writes.
type CompleteQueryCallback func(ctx context.Context) error
