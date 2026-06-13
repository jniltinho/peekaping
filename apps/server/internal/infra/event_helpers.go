package infra

import (
	"encoding/json"
	"peekaping/internal/modules/events"
)

// UnmarshalEventPayload extracts a typed *T from event.Payload. It handles
// both direct in-process type assertions and the json.RawMessage payloads
// produced by RedisEventBus during cross-process delivery.
func UnmarshalEventPayload[T any](event events.Event) (*T, bool) {
	if payload, ok := event.Payload.(*T); ok {
		return payload, true
	}

	// Redis serialises payloads to json.RawMessage; try that path too.
	if rawMsg, ok := event.Payload.(json.RawMessage); ok {
		var result T
		if err := json.Unmarshal(rawMsg, &result); err == nil {
			return &result, true
		}
	}

	return nil, false
}

// UnmarshalEventPayloadValue is like UnmarshalEventPayload but returns T by
// value instead of a pointer, useful when T is not pointer-sized or when
// the caller prefers value semantics.
func UnmarshalEventPayloadValue[T any](event events.Event) (T, bool) {
	var zero T

	if payload, ok := event.Payload.(T); ok {
		return payload, true
	}

	if payload, ok := event.Payload.(*T); ok {
		return *payload, true
	}

	if rawMsg, ok := event.Payload.(json.RawMessage); ok {
		var result T
		if err := json.Unmarshal(rawMsg, &result); err == nil {
			return result, true
		}
	}

	return zero, false
}
