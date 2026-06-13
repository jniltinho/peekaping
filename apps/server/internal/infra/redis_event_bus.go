package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"peekaping/internal/config"
	"peekaping/internal/modules/events"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisEventChannelPrefix is the namespace prefix for all Redis pub/sub channels.
const RedisEventChannelPrefix = "monitoring:events:"

// RedisEventBus implements events.EventBus using Redis pub/sub so events are
// delivered across multiple processes. Each EventType maps to one channel.
// A dedicated publishing client is maintained to avoid blocking the subscriber.
type RedisEventBus struct {
	client      *redis.Client
	pubClient   *redis.Client // separate client so Publish never blocks Subscribe
	logger      *zap.SugaredLogger
	mu          sync.RWMutex
	handlers    map[events.EventType][]events.EventHandler
	subscribers map[events.EventType]*redis.PubSub
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// SerializedEvent is the wire format carried over Redis channels.
// Payload is kept as raw JSON so each subscriber can decode into its own type.
type SerializedEvent struct {
	Type    events.EventType `json:"type"`
	Payload json.RawMessage  `json:"payload"`
}

// NewRedisEventBus constructs a RedisEventBus. The publish client is created
// from client.Options so it shares the same connection parameters.
func NewRedisEventBus(client *redis.Client, logger *zap.SugaredLogger) *RedisEventBus {
	ctx, cancel := context.WithCancel(context.Background())
	pubClient := redis.NewClient(client.Options())

	return &RedisEventBus{
		client:      client,
		pubClient:   pubClient,
		logger:      logger.With("component", "redis_event_bus"),
		handlers:    make(map[events.EventType][]events.EventHandler),
		subscribers: make(map[events.EventType]*redis.PubSub),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Subscribe registers handler for eventType. The first handler for a given
// event type starts a background goroutine that listens on the Redis channel.
func (b *RedisEventBus) Subscribe(eventType events.EventType, handler events.EventHandler) {
	b.logger.Debugf("Subscribing to event: %s", eventType)
	b.mu.Lock()
	defer b.mu.Unlock()

	handlers := b.handlers[eventType]
	handlers = append(handlers, handler)
	b.handlers[eventType] = handlers

	if len(handlers) == 1 {
		b.startRedisSubscription(eventType)
	}
}

// startRedisSubscription opens a Redis pub/sub channel for eventType and
// dispatches incoming messages to all registered handlers in separate goroutines.
func (b *RedisEventBus) startRedisSubscription(eventType events.EventType) {
	channel := b.getChannelName(eventType)
	pubsub := b.client.Subscribe(b.ctx, channel)
	b.subscribers[eventType] = pubsub

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.logger.Infof("Started Redis subscription for channel: %s", channel)

		if _, err := pubsub.Receive(b.ctx); err != nil {
			b.logger.Errorw("Failed to receive subscription confirmation",
				"channel", channel,
				"error", err,
			)
			return
		}

		ch := pubsub.Channel()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					b.logger.Infof("Channel closed for event type: %s", eventType)
					return
				}
				b.handleRedisMessage(eventType, msg)

			case <-b.ctx.Done():
				b.logger.Infof("Context cancelled, stopping subscription for: %s", eventType)
				return
			}
		}
	}()
}

// handleRedisMessage deserialises a raw Redis message and fans it out to all
// local handlers. Each handler runs in its own goroutine with panic recovery.
func (b *RedisEventBus) handleRedisMessage(eventType events.EventType, msg *redis.Message) {
	b.logger.Debugf("Received Redis message for event type: %s", eventType)

	var serialized SerializedEvent
	if err := json.Unmarshal([]byte(msg.Payload), &serialized); err != nil {
		b.logger.Errorw("Failed to unmarshal event",
			"event_type", eventType,
			"error", err,
		)
		return
	}

	event := events.Event{
		Type:    serialized.Type,
		Payload: serialized.Payload,
	}

	b.mu.RLock()
	handlers := b.handlers[eventType]
	b.mu.RUnlock()

	for _, handler := range handlers {
		go func(h events.EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					b.logger.Errorw("Event handler panicked",
						"event_type", eventType,
						"panic", r,
					)
				}
			}()
			h(event)
		}(handler)
	}
}

// Publish serialises event to JSON and delivers it to the corresponding Redis
// channel. Errors are logged but not returned; callers must not rely on
// delivery guarantees.
func (b *RedisEventBus) Publish(event events.Event) {
	b.logger.Debugf("Publishing event: %s", event.Type)

	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		b.logger.Errorw("Failed to marshal event payload",
			"event_type", event.Type,
			"error", err,
		)
		return
	}

	data, err := json.Marshal(SerializedEvent{
		Type:    event.Type,
		Payload: payloadJSON,
	})
	if err != nil {
		b.logger.Errorw("Failed to marshal event",
			"event_type", event.Type,
			"error", err,
		)
		return
	}

	channel := b.getChannelName(event.Type)
	if err = b.pubClient.Publish(b.ctx, channel, data).Err(); err != nil {
		b.logger.Errorw("Failed to publish event to Redis",
			"event_type", event.Type,
			"channel", channel,
			"error", err,
		)
	}

	b.logger.Debugf("Successfully published event to Redis: %s", event.Type)
}

// Close cancels all Redis subscriptions, waits for goroutines to exit, and
// closes the publish client.
func (b *RedisEventBus) Close() error {
	b.logger.Info("Closing Redis event bus")

	b.cancel()

	b.mu.Lock()
	for eventType, pubsub := range b.subscribers {
		if err := pubsub.Close(); err != nil {
			b.logger.Errorw("Failed to close subscription",
				"event_type", eventType,
				"error", err,
			)
		}
	}
	b.subscribers = make(map[events.EventType]*redis.PubSub)
	b.mu.Unlock()

	b.wg.Wait()

	if err := b.pubClient.Close(); err != nil {
		b.logger.Errorw("Failed to close pub client", "error", err)
	}

	b.logger.Info("Redis event bus closed")
	return nil
}

// getChannelName returns the full Redis channel name for the given event type.
func (b *RedisEventBus) getChannelName(eventType events.EventType) string {
	return fmt.Sprintf("%s%s", RedisEventChannelPrefix, eventType)
}

// GetStats returns a snapshot of handler and subscription counts keyed by
// event type, suitable for observability endpoints.
func (b *RedisEventBus) GetStats() map[string]any {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlerCounts := make(map[events.EventType]int, len(b.handlers))
	for eventType, handlers := range b.handlers {
		handlerCounts[eventType] = len(handlers)
	}

	return map[string]any{
		"total_event_types":   len(b.handlers),
		"total_subscriptions": len(b.subscribers),
		"handler_counts":      handlerCounts,
	}
}

// ProvideRedisClient constructs a *redis.Client from cfg for use in the DI container.
func ProvideRedisClient(cfg *config.Config, logger *zap.SugaredLogger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	logger.Info("Successfully created Redis client")
	return client, nil
}

// ProvideRedisEventBus wraps NewRedisEventBus and satisfies the events.EventBus
// interface for the DI container.
func ProvideRedisEventBus(client *redis.Client, logger *zap.SugaredLogger) events.EventBus {
	logger.Info("Creating Redis-based event bus")
	return NewRedisEventBus(client, logger)
}
