package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	redisJobQueueKey    = "peekaping:engine:jobs"
	redisResultQueueKey = "peekaping:engine:results"
)

// RedisQueue is a Redis-backed job queue using LPUSH/BRPOP for FIFO ordering.
type RedisQueue struct {
	client *redis.Client
	key    string
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client, key: redisJobQueueKey}
}

func (q *RedisQueue) Push(ctx context.Context, job CheckJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	return q.client.LPush(ctx, q.key, data).Err()
}

func (q *RedisQueue) Pop(ctx context.Context) (CheckJob, error) {
	result, err := q.client.BRPop(ctx, 0, q.key).Result()
	if err != nil {
		return CheckJob{}, fmt.Errorf("brpop job: %w", err)
	}
	var job CheckJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return CheckJob{}, fmt.Errorf("unmarshal job: %w", err)
	}
	return job, nil
}

// Close is a no-op; the client lifecycle is managed by the DI container.
func (q *RedisQueue) Close() {}

// RedisResultQueue is a Redis-backed result queue using LPUSH/BRPOP for FIFO ordering.
type RedisResultQueue struct {
	client *redis.Client
	key    string
}

func NewRedisResultQueue(client *redis.Client) *RedisResultQueue {
	return &RedisResultQueue{client: client, key: redisResultQueueKey}
}

func (q *RedisResultQueue) Push(ctx context.Context, result CheckResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	return q.client.LPush(ctx, q.key, data).Err()
}

func (q *RedisResultQueue) Pop(ctx context.Context) (CheckResult, error) {
	res, err := q.client.BRPop(ctx, 0, q.key).Result()
	if err != nil {
		return CheckResult{}, fmt.Errorf("brpop result: %w", err)
	}
	var result CheckResult
	if err := json.Unmarshal([]byte(res[1]), &result); err != nil {
		return CheckResult{}, fmt.Errorf("unmarshal result: %w", err)
	}
	return result, nil
}

// Close is a no-op; the client lifecycle is managed by the DI container.
func (q *RedisResultQueue) Close() {}
