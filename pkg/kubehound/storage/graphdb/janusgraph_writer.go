package graphdb

import (
	"context"
	"errors"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/path"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

type TWriterInput interface {
	vertex.Traversal | edge.Traversal | path.Traversal
}

type JanusGraphAsyncWriter[T TWriterInput] struct {
	label           string
	gremlin         T
	traversalSource *gremlingo.GraphTraversalSource
	inserts         []types.TraversalInput
	consumerChan    chan []types.TraversalInput
	writingInFlight *sync.WaitGroup
	batchSize       int
	mu              sync.Mutex
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

	op := jgv.gremlin(jgv.traversalSource, data)
	promise := op.Iterate()
	err := <-promise
	if err != nil {
		return err
	}

	return nil
}

func (jgv *JanusGraphAsyncWriter[T]) Close(ctx context.Context) error {
	close(jgv.consumerChan)
	return nil
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

	return nil
}

func (jgv *JanusGraphAsyncWriter[T]) Queue(ctx context.Context, v any) error {
	jgv.mu.Lock()
	defer jgv.mu.Unlock()

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
