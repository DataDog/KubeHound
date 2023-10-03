//nolint:all
package statsd

import (
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

var _ statsd.ClientInterface = &NoopClient{}

// NoopClient may be used in place of any type of statsd client that speaks statsd.ClientInterface
// for testing, to remove side-effects of statsd calls
type NoopClient struct{}

// NewNoopClient returns a statsd client implementation that does nothing. It is used for tests.
func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

func (*NoopClient) Count(name string, value int64, tags []string, rate float64) error {
	return nil
}

func (*NoopClient) CountWithTimestamp(name string, value int64, tags []string, rate float64, timestamp time.Time) error {
	return nil
}

func (*NoopClient) Gauge(name string, value float64, tags []string, rate float64) error {
	return nil
}

func (*NoopClient) GaugeWithTimestamp(name string, value float64, tags []string, rate float64, timestamp time.Time) error {
	return nil
}

func (*NoopClient) Incr(name string, tags []string, rate float64) error {
	return nil
}

func (*NoopClient) Decr(name string, tags []string, rate float64) error {
	return nil
}

func (*NoopClient) Histogram(name string, value float64, tags []string, rate float64) error {
	return nil
}

func (*NoopClient) Event(event *statsd.Event) error {
	return nil
}

func (*NoopClient) ServiceCheck(sc *statsd.ServiceCheck) error {
	return nil
}

func (*NoopClient) Set(name string, value string, tags []string, rate float64) error {
	return nil
}

func (*NoopClient) Timing(name string, value time.Duration, tags []string, rate float64) error {
	return nil
}

func (*NoopClient) Distribution(name string, value float64, tags []string, rate float64) error {
	return nil
}

func (*NoopClient) Close() error {
	return nil
}

func (*NoopClient) Flush() error {
	return nil
}

func (*NoopClient) GetTelemetry() statsd.Telemetry {
	return statsd.Telemetry{}
}

func (*NoopClient) IsClosed() bool {
	return false
}

func (*NoopClient) SimpleEvent(title, text string) error {
	return nil
}

func (*NoopClient) SimpleServiceCheck(name string, status statsd.ServiceCheckStatus) error {
	return nil
}

func (*NoopClient) TimeInMilliseconds(name string, value float64, tags []string, rate float64) error {
	return nil
}
