package graphdb

import (
	"context"
	"errors"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

var _ AsyncVertexWriter = (*JanusGraphAsyncVertexWriter)(nil)

type JanusGraphAsyncVertexWriter struct {
	gremlin              vertex.VertexTraversal
	transaction          *gremlingo.Transaction
	traversalSource      *gremlingo.GraphTraversalSource
	inserts              []vertex.TraversalInput
	consumerChan         chan []vertex.TraversalInput
	writingInFligth      sync.WaitGroup
	batchSize            int
	mu                   sync.Mutex
	isTransactionEnabled bool
}

func NewJanusGraphAsyncVertexWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection, v vertex.Builder, opts ...WriterOption) (*JanusGraphAsyncVertexWriter, error) {
	log.I.Infof("Created new JanusGraphAsyncVertexWriter")
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

	jw := JanusGraphAsyncVertexWriter{
		gremlin:         v.Traversal(),
		inserts:         make([]vertex.TraversalInput, 0, v.BatchSize()),
		transaction:     tx,
		traversalSource: source,
		batchSize:       v.BatchSize(),
		consumerChan:    make(chan []vertex.TraversalInput, v.BatchSize()*channelSizeBatchFactor),
	}
	jw.startBackgroundWriter(ctx)
	return &jw, nil
}

// startBackgroundWriter starts a background go routine
func (jgv *JanusGraphAsyncVertexWriter) startBackgroundWriter(ctx context.Context) {
	go func() {
		for {
			select {
			case data := <-jgv.consumerChan:
				// closing the channel shoud stop the go routine
				if data == nil {
					return
				}
				jgv.writingInFligth.Add(1)
				err := jgv.batchWrite(ctx, data)
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

func (jgv *JanusGraphAsyncVertexWriter) batchWrite(ctx context.Context, data []vertex.TraversalInput) error {
	log.I.Debugf("batch write JanusGraphAsyncVertexWriter with %d elements", len(data))
	defer jgv.writingInFligth.Done()

	// convertedToTraversalInput := make([]vertex.TraversalInput, 0)
	// for _, d := range data {
	// 	convertedToTraversalInput = append(convertedToTraversalInput, d.(vertex.TraversalInput))
	// }
	op := jgv.gremlin(jgv.traversalSource, data)
	promise := op.Iterate()
	err := <-promise
	if err != nil {
		if jgv.isTransactionEnabled {
			jgv.transaction.Rollback()
		}
		return err
	}
	if jgv.isTransactionEnabled {
		log.I.Infof("commiting work")
		err = jgv.transaction.Commit()
		if err != nil {
			log.I.Errorf("failed to commit: %+v", err)
			return err
		}
	}
	return nil
}

func (jgv *JanusGraphAsyncVertexWriter) Close(ctx context.Context) error {
	if jgv.isTransactionEnabled {
		return jgv.transaction.Close()
	}
	return nil
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (jgv *JanusGraphAsyncVertexWriter) Flush(ctx context.Context) error {
	jgv.mu.Lock()
	defer jgv.mu.Unlock()

	if jgv.traversalSource == nil {
		return errors.New("JanusGraph traversalSource is not initialized")
	}

	if len(jgv.inserts) == 0 {
		log.I.Debugf("Skipping flush on vertex writer as no write operations left")
		// we need to send something to the channel from this function whenever we don't return an error
		// we cannot defer it because the go routine may last longer than the current function
		// the defer is going to be executed at the return time, whetever or not the inner go routine is processing data
		jgv.writingInFligth.Wait()
		return nil
	}

	jgv.writingInFligth.Add(1)
	err := jgv.batchWrite(ctx, jgv.inserts)
	if err != nil {
		log.I.Errorf("Failed to batch write vertex: %+v", err)
		jgv.writingInFligth.Wait()
		return err
	}
	log.I.Info("Done flushing vertices, clearing the queue")
	jgv.inserts = nil

	return nil
}

func (jgv *JanusGraphAsyncVertexWriter) Queue(ctx context.Context, v any) error {
	jgv.mu.Lock()
	defer jgv.mu.Unlock()

	jgv.inserts = append(jgv.inserts, v)
	if len(jgv.inserts) > jgv.batchSize {
		var copied []vertex.TraversalInput
		copied = make([]vertex.TraversalInput, len(jgv.inserts))
		copy(copied, jgv.inserts)
		jgv.consumerChan <- copied
		// cleanup the ops array after we have copied it to the channel
		jgv.inserts = nil
	}
	return nil
}
