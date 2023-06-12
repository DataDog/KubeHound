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
	inserts         []any
	consumerChan    chan []any
	writingInFligth sync.WaitGroup
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
		inserts:         make([]any, 0, e.BatchSize()),
		traversalSource: source,
		batchSize:       e.BatchSize(),
		consumerChan:    make(chan []any, e.BatchSize()*channelSizeBatchFactor),
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
				jge.writingInFligth.Add(1)
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

func (jge *JanusGraphAsyncEdgeWriter) batchWrite(ctx context.Context, data []any) error {
	log.I.Debugf("batch write JanusGraphAsyncEdgeWriter with %d elements", len(data))
	defer jge.writingInFligth.Done()

	// This seems ~pointless BUT is required to have the ability to use edge.TraversalInput/vertex.TraversalInput
	// as the type
	// Even tho it's an alias to any, since we use it in an array, we cannot simply .([]any) or vice versa because of the underlying memory layout.
	convertedToTraversalInput := make([]edge.TraversalInput, 0)
	for _, d := range data {
		convertedToTraversalInput = append(convertedToTraversalInput, d.(edge.TraversalInput))
	}

	op := jge.gremlin(jge.traversalSource, convertedToTraversalInput)
	promise := op.Iterate()
	err := <-promise
	if err != nil {
		return err
	}

	return nil
}

func (jge *JanusGraphAsyncEdgeWriter) Close(ctx context.Context) error {
	return nil
}

func (jge *JanusGraphAsyncEdgeWriter) Flush(ctx context.Context) error {
	jge.mu.Lock()
	defer jge.mu.Unlock()

	if jge.traversalSource == nil {
		return errors.New("JanusGraph traversalSource is not initialized")
	}

	if len(jge.inserts) == 0 {
		log.I.Debugf("Skipping flush on edges writer as no write operations left")
		// we need to send something to the channel from this function whenever we don't return an error
		// we cannot defer it because the go routine may last longer than the current function
		// the defer is going to be executed at the return time, whetever or not the inner go routine is processing data
		jge.writingInFligth.Wait()
		return nil
	}

	jge.writingInFligth.Add(1)
	err := jge.batchWrite(ctx, jge.inserts)
	if err != nil {
		log.I.Errorf("Failed to batch write edge: %+v", err)
		jge.writingInFligth.Wait()
		return err
	}
	log.I.Info("Done flushing edges, clearing the queue")
	jge.inserts = nil

	return nil
}

func (jge *JanusGraphAsyncEdgeWriter) Queue(ctx context.Context, edge any) error {
	jge.mu.Lock()
	defer jge.mu.Unlock()

	jge.inserts = append(jge.inserts, edge)
	if len(jge.inserts) > jge.batchSize {
		copied := make([]any, len(jge.inserts))
		copy(copied, jge.inserts)
		jge.consumerChan <- copied
		// cleanup the ops array after we have copied it to the channel
		jge.inserts = nil
	}
	return nil
}
