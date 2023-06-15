package storedb

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
)

type writerOptions struct {
}

type WriterOption func(*writerOptions)

// Provider defines the interface for implementations of the storedb provider for intermediate storage of normalized K8s data.
//
//go:generate mockery --name Provider --output mocks --case underscore --filename store_provider.go --with-expecter
type Provider interface {
	services.Dependency

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

	// Flush triggers writes of any remaining items in the queue. Blocks until operation completes.
	Flush(ctx context.Context) error

	// Close cleans up any resources used by the AsyncWriter implementation. Writer cannot be reused after this call.
	Close(ctx context.Context) error
}

// Factory returns an initialized instance of a storedb provider from the provided application config.
func Factory(ctx context.Context, cfg *config.KubehoundConfig) (Provider, error) {
	r := storage.Retrier(NewMongoProvider, cfg.Storage.Retry, cfg.Storage.RetryDelay)
	return r(ctx, cfg.MongoDB.URL, cfg.MongoDB.ConnectionTimeout)
}
