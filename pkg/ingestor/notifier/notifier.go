package notifier

import "context"

type Notifier interface {
	// Notify notifies for the completion of ingestion of a cluster and run ID
	// Example use case can be a queuing system, a webhook, or anything else.
	Notify(ctx context.Context, clusterName string, runID string) error
}

//go:generate mockery --name=Notifier --output=mocks --outpkg=mocks --case=underscore
