package queue

import (
	"context"
	"sync"
	"time"
)

// FlushFunc is called when queue needs to be flushed to persistence
type FlushFunc[T any] func(ctx context.Context, items []T) error

// MemoryBatchQueue is a generic in-memory batch queue
// It supports flush on reaching batch size or after a time interval.
type MemoryBatchQueue[T any] struct {
	mu        sync.Mutex
	batch     []T
	batchSize int
	interval  time.Duration
	flushFunc FlushFunc[T]
	stopCh    chan struct{}
}

// NewMemoryBatchQueue creates a new in-memory batch queue.
// - batchSize: number of items to accumulate before flush
// - interval: max time to wait before flushing (even if batch not full)
func NewMemoryBatchQueue[T any](batchSize int, interval time.Duration, flushFunc FlushFunc[T]) *MemoryBatchQueue[T] {
	q := &MemoryBatchQueue[T]{
		batch:     make([]T, 0, batchSize),
		batchSize: batchSize,
		interval:  interval,
		flushFunc: flushFunc,
		stopCh:    make(chan struct{}),
	}
	go q.runFlusher()
	return q
}

// Enqueue adds an item to the queue and triggers flush if batch size is reached
func (q *MemoryBatchQueue[T]) Enqueue(ctx context.Context, item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.batch = append(q.batch, item)
	if len(q.batch) >= q.batchSize {
		return q.flushLocked(ctx)
	}
	return nil
}

// Flush forces flushing of all pending items
func (q *MemoryBatchQueue[T]) Flush(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.flushLocked(ctx)
}

// Stop stops the background flusher goroutine
func (q *MemoryBatchQueue[T]) Stop() {
	close(q.stopCh)
}

// runFlusher periodically flushes the queue by interval
func (q *MemoryBatchQueue[T]) runFlusher() {
	ticker := time.NewTicker(q.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = q.Flush(context.Background()) // ignore error here, service layer can handle retries
		case <-q.stopCh:
			return
		}
	}
}

// flushLocked assumes mutex is already held
func (q *MemoryBatchQueue[T]) flushLocked(ctx context.Context) error {
	if len(q.batch) == 0 {
		return nil
	}

	items := q.batch
	q.batch = make([]T, 0, q.batchSize)

	err := q.flushFunc(ctx, items)
	if err != nil {
		// restore items if flush failed
		q.batch = append(q.batch, items...)
		return err
	}
	return nil
}
