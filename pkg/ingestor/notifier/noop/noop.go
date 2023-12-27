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
	log.I.Warnf("Noop Notifying for cluster %s and run ID %s", clusterName, runID)
	return nil
}
