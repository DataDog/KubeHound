package storedb

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
)

// Provider defines the interface for implementations of the storedb provider for intermediate storage of normalized K8s data.
//
//go:generate mockery --name Provider --output mocks --case underscore --filename store_provider.go --with-expecter
type Provider interface {
	services.Dependency

	// Prepare drops and recreates all tables and indices (idempotent reset).
	Prepare(ctx context.Context) error

	// Reader returns a handle to the underlying provider to allow implementation specific queries against the store DB
	Reader() any

	// Write performs a synchronous insert of a store model.
	Write(ctx context.Context, model any) error

	// Close cleans up any resources used by the Provider implementation. Provider cannot be reused after this call.
	Close(ctx context.Context) error
}

// Factory returns an initialized instance of a storedb provider from the provided application config.
func Factory(ctx context.Context, cfg *config.KubehoundConfig) (Provider, error) {
	return NewSQLiteProvider(ctx, cfg)
}
