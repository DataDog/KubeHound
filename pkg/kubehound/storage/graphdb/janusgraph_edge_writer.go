package graphdb

import (
	"context"
	"errors"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

var _ AsyncEdgeWriter = (*JanusGraphAsyncEdgeWriter)(nil)

type JanusGraphAsyncEdgeWriter struct {
	gremlin         edge.EdgeTraversal
	traversalSource *gremlingo.GraphTraversalSource
	inserts         []edge.TraversalInput
	consumerChan    chan []edge.TraversalInput
	writingInFlight *sync.WaitGroup
	batchSize       int
	mu              sync.Mutex
}

func NewJanusGraphAsyncEdgeWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection, e edge.Builder, opts ...WriterOption) (*JanusGraphAsyncEdgeWriter, error) {
	log.I.Infof("Created new JanusGraphAsyncEdgeWriter")
	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	source := gremlingo.Traversal_().WithRemote(drc)
	jw := JanusGraphAsyncEdgeWriter{
		gremlin:         e.Traversal(),
		inserts:         make([]edge.TraversalInput, 0, e.BatchSize()),
		traversalSource: source,
		batchSize:       e.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan []edge.TraversalInput, e.BatchSize()*channelSizeBatchFactor),
	}
	jw.startBackgroundWriter(ctx)
	return &jw, nil
}

// startBackgroundWriter starts a background go routine
func (jge *JanusGraphAsyncEdgeWriter) startBackgroundWriter(ctx context.Context) {
	go func() {
		for {
			select {
			case data := <-jge.consumerChan:
				// closing the channel shoud stop the go routine
				if data == nil {
					return
				}
				err := jge.batchWrite(ctx, data)
				if err != nil {
					log.I.Errorf("failed to write data in background batch writer: %v", err)
				}
			case <-ctx.Done():
				log.I.Info("Closed background janusgraph worker")
				return
			}
		}
	}()
}

func (jge *JanusGraphAsyncEdgeWriter) batchWrite(ctx context.Context, data []edge.TraversalInput) error {
	log.I.Debugf("batch write JanusGraphAsyncEdgeWriter with %d elements", len(data))
	defer jge.writingInFlight.Done()

	op := jge.gremlin(jge.traversalSource, data)
	promise := op.Iterate()
	err := <-promise
	if err != nil {
		return err
	}

	return nil
}

func (jge *JanusGraphAsyncEdgeWriter) Close(ctx context.Context) error {
	close(jge.consumerChan)
	return nil
}

func (jge *JanusGraphAsyncEdgeWriter) Flush(ctx context.Context) error {
	jge.mu.Lock()
	defer jge.mu.Unlock()

	if jge.traversalSource == nil {
		return errors.New("JanusGraph traversalSource is not initialized")
	}

	if len(jge.inserts) != 0 {
		jge.writingInFlight.Add(1)
		err := jge.batchWrite(ctx, jge.inserts)
		if err != nil {
			log.I.Errorf("Failed to batch write edge: %+v", err)
			jge.writingInFlight.Wait()
			return err
		}
		log.I.Info("Done flushing edges, clearing the queue")
		jge.inserts = nil
	}

	jge.writingInFlight.Wait()

	return nil
}

func (jge *JanusGraphAsyncEdgeWriter) Queue(ctx context.Context, e any) error {
	jge.mu.Lock()
	defer jge.mu.Unlock()

	jge.inserts = append(jge.inserts, e)
	if len(jge.inserts) > jge.batchSize {
		copied := make([]edge.TraversalInput, len(jge.inserts))
		copy(copied, jge.inserts)
		jge.writingInFlight.Add(1)
		jge.consumerChan <- copied
		// cleanup the ops array after we have copied it to the channel
		jge.inserts = nil
	}
	return nil
}
