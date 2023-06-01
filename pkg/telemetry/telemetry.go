package telemetry

import (
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/datadog-go/v5/statsd"
)

// just to make sure we have a client that does nothing by default
func init() {
	statsdClient = &NoopClient{}
}

func Initialize(statsdURL string) error {
	var err error
	// In case we don't have a statsd url set, we just want to continue, but log that we aren't going to submit metrics.
	if statsdURL == "" {
		log.I.Warn("No metrics collector has been setup. All metrics submission are going to be NOOP.")
		return nil
	}
	statsdClient, err = statsd.New(statsdURL)
	if err != nil {
		return err
	}
	// Initialize all telemetry required
	// return client to enable clean shutdown
	return nil
}

func Shutdown() {
	err := statsdClient.Flush()
	log.I.Warnf("Failed to flush statsd client: %v", err)
	statsdClient.Close()
	log.I.Warnf("Failed to close statsd client: %v", err)
}
