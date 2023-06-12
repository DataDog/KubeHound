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
	gremlin              edge.EdgeTraversal
	transaction          *gremlingo.Transaction
	traversalSource      *gremlingo.GraphTraversalSource
	inserts              []any
	consumerChan         chan []any
	writingInFligth      sync.WaitGroup
	batchSize            int
	mu                   sync.Mutex
	isTransactionEnabled bool
}

func NewJanusGraphAsyncEdgeWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection, e edge.Builder, opts ...WriterOption) (*JanusGraphAsyncEdgeWriter, error) {
	log.I.Infof("Created new JanusGraphAsyncEdgeWriter")
	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	source := gremlingo.Traversal_().WithRemote(drc)
	// quick switch to enable / disable transaction
	var tx *gremlingo.Transaction
	if options.isTransactionEnabled {
		log.I.Info("GraphDB transaction enabled!")
		tx = source.Tx()
		var err error
		source, err = tx.Begin()
		if err != nil {
			return nil, err
		}
	}

	jw := JanusGraphAsyncEdgeWriter{
		gremlin:         e.Traversal(),
		inserts:         make([]any, 0),
		traversalSource: source,
		transaction:     tx,
		batchSize:       e.BatchSize(),
		consumerChan:    make(chan []any, e.BatchSize()*channelSizeBatchFactor),
	}
	jw.backgroundWriter(ctx)
	return &jw, nil
}

// backgroundWriter starts a background go routine
func (jge *JanusGraphAsyncEdgeWriter) backgroundWriter(ctx context.Context) {
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

func (jge *JanusGraphAsyncEdgeWriter) batchWrite(ctx context.Context, data []any) error {
	log.I.Debugf("batch write JanusGraphAsyncEdgeWriter with %d elements", len(data))
	jge.writingInFligth.Add(1)
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
		if jge.isTransactionEnabled {
			jge.transaction.Rollback()
		}
		return err
	}

	if jge.isTransactionEnabled {
		err = jge.transaction.Commit()
		if err != nil {
			log.I.Errorf("failed to commit: %+v", err)
			return err
		}
	}
	return nil
}

func (e *JanusGraphAsyncEdgeWriter) Close(ctx context.Context) error {
	if e.isTransactionEnabled {
		return e.transaction.Close()
	}
	return nil
}

func (e *JanusGraphAsyncEdgeWriter) Flush(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.traversalSource == nil {
		return errors.New("JanusGraph traversalSource is not initialized")
	}

	if len(e.inserts) == 0 {
		log.I.Debugf("Skipping flush on edges writer as no write operations left")
		// we need to send something to the channel from this function whenever we don't return an error
		// we cannot defer it because the go routine may last longer than the current function
		// the defer is going to be executed at the return time, whetever or not the inner go routine is processing data
		e.writingInFligth.Wait()
		return nil
	}
	err := e.batchWrite(ctx, e.inserts)
	if err != nil {
		log.I.Errorf("Failed to batch write edge: %+v", err)
		e.writingInFligth.Wait()
		return err
	}
	log.I.Info("Done flushing edges, clearing the queue")
	e.inserts = nil

	return nil
}

func (e *JanusGraphAsyncEdgeWriter) Queue(ctx context.Context, edge any) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.inserts) > e.batchSize {
		copied := make([]any, len(e.inserts))
		copy(copied, e.inserts)
		e.consumerChan <- copied
		// cleanup the ops array after we have copied it to the channel
		e.inserts = nil
	}
	e.inserts = append(e.inserts, edge)
	return nil
}
