package engine

import (
	"context"
	"fmt"
)

// RedisQueue is a placeholder for a future Redis-backed job queue (ENGINE_USE_REDIS=true).
// It is not yet implemented; use MemoryQueue for production deployments.
type RedisQueue struct{}

func (q *RedisQueue) Push(_ context.Context, _ CheckJob) error {
	return fmt.Errorf("RedisQueue not yet implemented; set ENGINE_USE_REDIS=false")
}

func (q *RedisQueue) Pop(_ context.Context) (CheckJob, error) {
	return CheckJob{}, fmt.Errorf("RedisQueue not yet implemented; set ENGINE_USE_REDIS=false")
}

func (q *RedisQueue) Close() {}

// RedisResultQueue is a placeholder for a future Redis-backed result queue.
type RedisResultQueue struct{}

func (q *RedisResultQueue) Push(_ context.Context, _ CheckResult) error {
	return fmt.Errorf("RedisResultQueue not yet implemented; set ENGINE_USE_REDIS=false")
}

func (q *RedisResultQueue) Pop(_ context.Context) (CheckResult, error) {
	return CheckResult{}, fmt.Errorf("RedisResultQueue not yet implemented; set ENGINE_USE_REDIS=false")
}

func (q *RedisResultQueue) Close() {}
