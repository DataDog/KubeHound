package storedb

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
)

type writerOptions struct {
}

type WriterOption func(*writerOptions)

// Provider defines the interface for implementations of the storedb provider for intermediate storage of normalized K8s data.
//
//go:generate mockery --name Provider --output mocks --case underscore --filename store_provider.go --with-expecter
type Provider interface {
	// HealthCheck provides a mechanism for the client to check health of the provider.
	// Should return true if health check is successful, false otherwise.
	HealthCheck(ctx context.Context) (bool, error)

	// Raw returns a handle to the underlying provider to allow implementation specific operations e.g db queries.
	Raw() any

	// BulkWriter creates a new AsyncWriter instance to enable asynchronous bulk inserts.
	BulkWriter(ctx context.Context, collection collections.Collection, opts ...WriterOption) (AsyncWriter, error)

	// Close cleans up any resources used by the Provider implementation. Provider cannot be reused after this call.
	Close(ctx context.Context) error
}

// AysncWriter defines the interface for writer clients to queue aysnchronous, batched writes to the storedb.
//
//go:generate mockery --name AsyncWriter --output mocks --case underscore --filename store_writer.go --with-expecter
type AsyncWriter interface {
	// Queue add a model to an asynchronous write queue. Non-blocking.
	Queue(ctx context.Context, model any) error

	// Flush triggers writes of any remaining items in the queue.
	// Blocks until operation completes. Wait on the returned channel which will be signaled when the flush operation completes.
	Flush(ctx context.Context) (chan struct{}, error)

	// Close cleans up any resources used by the AsyncWriter implementation. Writer cannot be reused after this call.
	Close(ctx context.Context) error
}

// Factory returns an initialized instance of a storedb provider from the provided application config.
func Factory(ctx context.Context, cfg *config.KubehoundConfig) (Provider, error) {
	return nil, globals.ErrNotImplemented
}
