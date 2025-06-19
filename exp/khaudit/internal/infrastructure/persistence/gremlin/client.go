package gremlin

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

// ErrTimeout is returned when a Gremlin query times out.
var ErrTimeout = errors.New("timeout error")

// connHandler is called to create a connection with a Gremlin server.
type connHandler func(cfg Config) (*gremlingo.DriverRemoteConnection, error)

// A Connection handles the connection with the Gremlin server. This includes
// authentication, reconnections and retries.
type Connection struct {
	cfg Config
	h   connHandler
}

// NewConnection creates a [Connection] with the provided configuration.
func NewConnection(cfg Config) (Connection, error) {
	var connHandler connHandler

	switch cfg.AuthMode {
	case "plain":
		connHandler = connectPlain
	default:
		return Connection{}, errors.New("invalid auth mode")
	}

	conn := Connection{
		cfg: cfg,
		h:   connHandler,
	}
	return conn, nil
}

// connectPlain is a [connHandler] for Gremlin server that creates an
// unauthenticated connection.
func connectPlain(cfg Config) (*gremlingo.DriverRemoteConnection, error) {
	conn, err := gremlingo.NewDriverRemoteConnection(cfg.Endpoint, func(settings *gremlingo.DriverRemoteConnectionSettings) {
		settings.LogVerbosity = gremlingo.Off
	})
	return conn, err
}

// QueryFunc represents a Gremlin query in the context of a [Connection]. It is
// executed by [Connection.Query].
type QueryFunc func(*gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error)

// Query executes cf taking care of the authentication, reconnections and
// retries.
func (conn Connection) Query(cf QueryFunc) (results []*gremlingo.Result, err error) {
	for i := 0; i < conn.cfg.RetryLimit+1; i++ {
		results, err = conn.execQuery(cf)
		if err == nil {
			return results, nil
		}

		if strings.Contains(err.Error(), `"code":"TimeLimitExceededException"`) {
			return nil, ErrTimeout
		}

		if i < conn.cfg.RetryLimit {
			jitter := time.Duration(rand.Int63n(1000)) * time.Millisecond
			t := conn.cfg.RetryDuration + jitter

			time.Sleep(t)
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", err)
}

// execQuery executes cf in the context of a new remote Gremlin connection.
func (conn Connection) execQuery(cf QueryFunc) ([]*gremlingo.Result, error) {
	rc, err := conn.h(conn.cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating driver remote connection: %v", err)
	}
	defer rc.Close()

	g := gremlingo.Traversal_().WithRemote(rc)
	return cf(g)
}
