package graphdb

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// batchItem is a single item in the batch writer queue that contains the data
// to be written and the number of retries.
type batchItem struct {
	data       []any
	retryCount int
}

// microBatcher is a utility to batch items and flush them when the batch is full.
type microBatcher struct {
	// batchSize is the maximum number of items to batch.
	batchSize int
	// items is the current item accumulator for the batch. This is reset after
	// the batch is flushed.
	items []any
	// flush is the function to call to flush the batch.
	flushFunc func(context.Context, []any) error
	// itemChan is the channel to receive items to batch.
	itemChan chan any
	// batchChan is the channel to send batches to.
	batchChan chan batchItem
	// workerCount is the number of workers to process the batch.
	workerCount int
	// workerGroup is the worker group to wait for the workers to finish.
	workerGroup *sync.WaitGroup
	// shuttingDown is a flag to indicate if the batcher is shutting down.
	shuttingDown atomic.Bool
	// logger is the logger to use for logging.
	logger log.LoggerI
}

// NewMicroBatcher creates a new micro batcher.
func newMicroBatcher(logger log.LoggerI, batchSize int, workerCount int, flushFunc func(context.Context, []any) error) *microBatcher {
	return &microBatcher{
		logger:      logger,
		batchSize:   batchSize,
		items:       make([]any, 0, batchSize),
		flushFunc:   flushFunc,
		itemChan:    make(chan any, batchSize),
		batchChan:   make(chan batchItem, batchSize),
		workerCount: workerCount,
		workerGroup: nil, // Set in Start.
	}
}

// Flush flushes the current batch and waits for the batch writer to finish.
func (mb *microBatcher) Flush(_ context.Context) error {
	// Set the shutting down flag to true.
	if !mb.shuttingDown.CompareAndSwap(false, true) {
		return errors.New("batcher is already shutting down")
	}

	// Closing the item channel to signal the accumulator to stop and flush the batch.
	close(mb.itemChan)

	// Wait for the workers to finish.
	if mb.workerGroup != nil {
		mb.workerGroup.Wait()
	}

	return nil
}

// Enqueue adds an item to the batch processor.
func (mb *microBatcher) Enqueue(ctx context.Context, item any) error {
	// If the batcher is shutting down, return an error immediately.
	if mb.shuttingDown.Load() {
		return errors.New("batcher is shutting down")
	}

	select {
	case <-ctx.Done():
		// If the context is cancelled, return.
		return ctx.Err()
	case mb.itemChan <- item:
	}

	return nil
}

// Start starts the batch processor.
func (mb *microBatcher) Start(ctx context.Context) {
	if mb.workerGroup != nil {
		// If the worker group is already set, return.
		return
	}

	var wg sync.WaitGroup

	// Start the workers.
	for i := 0; i < mb.workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := mb.worker(ctx, mb.batchChan); err != nil {
				mb.logger.Errorf("worker: %v", err)
			}
		}()
	}

	// Start the item accumulator.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mb.runItemBatcher(ctx); err != nil {
			mb.logger.Errorf("run item batcher: %v", err)
		}

		// Close the batch channel to signal the workers to stop.
		close(mb.batchChan)
	}()

	// Set the worker group to wait for the workers to finish.
	mb.workerGroup = &wg
}

// startItemBatcher starts the item accumulator to batch items.
func (mb *microBatcher) runItemBatcher(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case item, ok := <-mb.itemChan:
			if !ok {
				// If the item channel is closed, send the current batch and return.
				mb.batchChan <- batchItem{
					data:       mb.items,
					retryCount: 0,
				}

				// End the accumulator.
				return nil
			}

			// Add the item to the batch.
			mb.items = append(mb.items, item)

			// If the batch is full, send it.
			if len(mb.items) == mb.batchSize {
				// Send the batch to the processor.
				mb.batchChan <- batchItem{
					data:       mb.items,
					retryCount: 0,
				}

				// Reset the batch.
				mb.items = mb.items[len(mb.items):]
			}
		}
	}
}

// startWorkers starts the workers to process the batches.
func (mb *microBatcher) worker(ctx context.Context, batchQueue <-chan batchItem) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case batch, ok := <-batchQueue:
			if !ok {
				return nil
			}

			// Send the batch to the processor.
			if len(batch.data) > 0 && mb.flushFunc != nil {
				if err := mb.flushFunc(ctx, batch.data); err != nil {
					mb.logger.Errorf("flush data in background batch writer: %v", err)
				}
			}
		}
	}
}
