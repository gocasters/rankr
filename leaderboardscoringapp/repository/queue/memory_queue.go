package queue

import (
	"context"
	"sync"
)

// MemoryQueue is a simple in-memory queue for generic items.
// Responsibilities:
// - Enqueue items
// - Return all current items without removing them
// - Allow manual flush (clear) when service confirms persistence
type MemoryQueue[T any] struct {
	mu    sync.Mutex
	items []T
}

// NewMemoryQueue creates a new empty memory queue
func NewMemoryQueue[T any]() *MemoryQueue[T] {
	return &MemoryQueue[T]{
		items: make([]T, 0),
	}
}

// Enqueue adds an item to the queue
func (q *MemoryQueue[T]) Enqueue(ctx context.Context, item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = append(q.items, item)

	return nil
}

// GetAll returns a copy of all items currently in the queue (without clearing)
func (q *MemoryQueue[T]) GetAll() []T {
	q.mu.Lock()
	defer q.mu.Unlock()

	copied := make([]T, len(q.items))
	copy(copied, q.items)
	return copied
}

// Clear empties the queue. Service layer should call this
// only after successful persistence
func (q *MemoryQueue[T]) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = q.items[:0]
}

// Size returns the number of items currently in the queue
func (q *MemoryQueue[T]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}
