package gremlin

import "time"

// Config contains the configuration parameters needed to interact with a
// Gremlin server.
type Config struct {
	// Endpoint is the Gremlin Endpoint.
	Endpoint string
	// AuthMode is the authentication mode.
	AuthMode string
	// RetryLimit is the number of retries before returning error.
	RetryLimit int
	// RetryDuration is the time to wait between retries.
	RetryDuration time.Duration
}
