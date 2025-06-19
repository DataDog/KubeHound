package renderer

import (
	"context"
	"io"
)

// Renderer is the interface for rendering reports.
type Renderer interface {
	// Render renders a report into a writer for a given cluster and run ID.
	Render(ctx context.Context, writer io.Writer, cluster string, runID string) error
}
