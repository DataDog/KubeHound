package noop

import (
	"context"

	notifier "github.com/DataDog/KubeHound/pkg/ingestor/notifier"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

type NoopNotifier struct{}

func NewNoopNotifier() notifier.Notifier {
	return &NoopNotifier{}
}

func (n *NoopNotifier) Notify(ctx context.Context, clusterName string, runID string) error {
	l := log.Logger(ctx)
	l.Warn("Noop Notifying for cluster and run ID", log.String(log.FieldClusterKey, clusterName), log.String(log.FieldRunIDKey, runID))

	return nil
}
