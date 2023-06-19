package graphdb

import (
	"sync"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

// DriverConnectionPool wraps access to the gremlin driver remote connection
// NOTE: the current gremlingo implementation is not thread safe. Creating, committing, closing transactions
// will result in unsafe access to an array not protected by a lock. This aims to fix this in as lightweight
// a manner as possible by linking a mutex to the connection to be acquired when doing certain operations that
// can modify the connection pool.
// See: github.com/apache/tinkerpop/gremlin-go/v3@v3.6.4/driver/graphTraversal.go:786 for offending code
type DriverConnectionPool struct {
	Lock   *sync.Mutex
	Driver *gremlingo.DriverRemoteConnection
}
