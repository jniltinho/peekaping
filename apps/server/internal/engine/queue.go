package engine

import (
	"context"
	"fmt"
)

// MemoryQueue is a buffered in-process job queue backed by a Go channel.
type MemoryQueue struct {
	jobs chan CheckJob
}

// NewMemoryQueue creates a MemoryQueue with the given buffer size.
func NewMemoryQueue(bufferSize int) *MemoryQueue {
	return &MemoryQueue{jobs: make(chan CheckJob, bufferSize)}
}

func (q *MemoryQueue) Push(ctx context.Context, job CheckJob) error {
	select {
	case q.jobs <- job:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("push job cancelled: %w", ctx.Err())
	}
}

func (q *MemoryQueue) Pop(ctx context.Context) (CheckJob, error) {
	select {
	case job, ok := <-q.jobs:
		if !ok {
			return CheckJob{}, fmt.Errorf("job queue closed")
		}
		return job, nil
	case <-ctx.Done():
		return CheckJob{}, fmt.Errorf("pop job cancelled: %w", ctx.Err())
	}
}

func (q *MemoryQueue) Close() {
	close(q.jobs)
}

// MemoryResultQueue is a buffered in-process result queue backed by a Go channel.
type MemoryResultQueue struct {
	results chan CheckResult
}

// NewMemoryResultQueue creates a MemoryResultQueue with the given buffer size.
func NewMemoryResultQueue(bufferSize int) *MemoryResultQueue {
	return &MemoryResultQueue{results: make(chan CheckResult, bufferSize)}
}

func (q *MemoryResultQueue) Push(ctx context.Context, result CheckResult) error {
	select {
	case q.results <- result:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("push result cancelled: %w", ctx.Err())
	}
}

func (q *MemoryResultQueue) Pop(ctx context.Context) (CheckResult, error) {
	select {
	case result, ok := <-q.results:
		if !ok {
			return CheckResult{}, fmt.Errorf("result queue closed")
		}
		return result, nil
	case <-ctx.Done():
		return CheckResult{}, fmt.Errorf("pop result cancelled: %w", ctx.Err())
	}
}

func (q *MemoryResultQueue) Close() {
	close(q.results)
}
