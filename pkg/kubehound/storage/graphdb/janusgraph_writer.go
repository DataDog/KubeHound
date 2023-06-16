package graphdb

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/path"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type TWriterInput interface {
	vertex.Traversal | edge.Traversal | path.Traversal
}

type JanusGraphAsyncWriter[T TWriterInput] struct {
	label           string                          // Label of the graph entity being written
	gremlin         T                               // Gremlin traversal generator function
	dcp             *DriverConnectionPool           // Lock protected gremlin driver remote connection
	traversalSource *gremlingo.GraphTraversalSource // Transacted graph traversal source
	transaction     *gremlingo.Transaction          // Transaction holding all the writes done by this writer
	inserts         []types.TraversalInput          // Object data to be inserted in the graph
	mu              sync.Mutex                      // Mutex protecting access to the inserts array
	consumerChan    chan []types.TraversalInput     // Channel consuming inserts for async writing
	writingInFlight *sync.WaitGroup                 // Wait group tracking current unfinished writes
	batchSize       int                             // Batchsize of graph DB inserts
	qcounter        int32                           // Track items queued
	wcounter        int32                           // Track items writtn
}

// startBackgroundWriter starts a background go routine
func (jgv *JanusGraphAsyncWriter[T]) startBackgroundWriter(ctx context.Context) {
	go func() {
		for {
			select {
			case data := <-jgv.consumerChan:
				// closing the channel shoud stop the go routine
				if data == nil {
					return
				}
				err := jgv.batchWrite(ctx, data)
				if err != nil {
					log.I.Errorf("write data in background batch writer: %v", err)
				}
			case <-ctx.Done():
				log.I.Info("Closed background janusgraph worker on context cancel")
				return
			}
		}
	}()
}

// batchWrite will write a batch of entries into the graph DB and block until the write completes.
// Callers are responsible for doing an Add(1) to the writingInFlight wait group to ensure proper synchronization.
func (jgv *JanusGraphAsyncWriter[T]) batchWrite(ctx context.Context, data []types.TraversalInput) error {
	log.I.Debugf("batch write JanusGraphAsyncVertexWriter with %d elements", len(data))
	defer jgv.writingInFlight.Done()

	atomic.AddInt32(&jgv.wcounter, int32(len(data)))

	op := jgv.gremlin(jgv.traversalSource, data)
	promise := op.Iterate()
	err := <-promise
	if err != nil {
		jgv.transaction.Rollback()
		return err
	}

	return nil
}

func (jgv *JanusGraphAsyncWriter[T]) Close(ctx context.Context) error {
	close(jgv.consumerChan)

	// Closing a transaction modifies the connection pool, acquire the lock
	jgv.dcp.Lock.Lock()
	defer jgv.dcp.Lock.Unlock()

	return jgv.transaction.Close()
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (jgv *JanusGraphAsyncWriter[T]) Flush(ctx context.Context) error {
	jgv.mu.Lock()
	defer jgv.mu.Unlock()

	if jgv.traversalSource == nil {
		return errors.New("JanusGraph traversalSource is not initialized")
	}

	if len(jgv.inserts) != 0 {
		jgv.writingInFlight.Add(1)
		err := jgv.batchWrite(ctx, jgv.inserts)
		if err != nil {
			log.I.Errorf("batch write %s: %+v", jgv.label, err)
			jgv.writingInFlight.Wait()
			return err
		}

		log.I.Infof("Done flushing %s writes. clearing the queue", jgv.label)
		jgv.inserts = nil
	}

	jgv.writingInFlight.Wait()

	// Committing a transaction modifies the connection pool, acquire the lock
	jgv.dcp.Lock.Lock()
	defer jgv.dcp.Lock.Unlock()

	err := jgv.transaction.Commit()
	if err != nil {
		return err
	}

	// TODO replace with metrics
	log.I.Infof("%d %s queued", jgv.qcounter, jgv.label)
	log.I.Infof("%d %s written", jgv.wcounter, jgv.label)
	return nil
}

func (jgv *JanusGraphAsyncWriter[T]) Queue(ctx context.Context, v any) error {
	jgv.mu.Lock()
	defer jgv.mu.Unlock()

	atomic.AddInt32(&jgv.qcounter, 1)
	jgv.inserts = append(jgv.inserts, v)
	if len(jgv.inserts) > jgv.batchSize {
		copied := make([]types.TraversalInput, len(jgv.inserts))
		copy(copied, jgv.inserts)

		jgv.writingInFlight.Add(1)
		jgv.consumerChan <- copied

		// cleanup the ops array after we have copied it to the channel
		jgv.inserts = nil
	}

	return nil
}
