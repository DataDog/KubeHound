package services

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// Dependency provides a unified interface for a service dependency requiring a health check.
type Dependency interface {
	// Name returns the name of the service dependency
	Name() string

	// HealthCheck provides a mechanism for the client to check health of the provider.
	// Should return true if health check is successful, false otherwise.
	HealthCheck(ctx context.Context) (bool, error)
}

// HealthCheck performs a health check on each provided dependency and returns an error on any failures.
func HealthCheck(ctx context.Context, deps []Dependency) error {
	var res *multierror.Error

	for _, d := range deps {
		ok, err := d.HealthCheck(ctx)
		if err != nil {
			res = multierror.Append(res, err)
		}

		if !ok {
			res = multierror.Append(res, fmt.Errorf("%s healthcheck", d.Name()))
		}
	}

	return res.ErrorOrNil()
}
