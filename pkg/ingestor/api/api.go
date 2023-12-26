package api

import "context"

type API interface {
	Ingest(ctx context.Context, clusterName string, runID string) error
	Listen(ctx context.Context) error
}
