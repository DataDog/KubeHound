package graphdb

import (
	"sync"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type DriverConnectionPool struct {
	Lock   *sync.Mutex
	Driver *gremlingo.DriverRemoteConnection
}
