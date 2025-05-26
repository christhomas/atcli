package services

import (
	"atcli/src/types"
	"reflect"
	"sync"
)

// EventBus manages event subscriptions and publishing
type EventBus struct {
	subscribers map[types.EventType][]types.EventHandlerFunc
	lock        sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[types.EventType][]types.EventHandlerFunc),
	}
}

// Subscribe registers a handler for a specific event type
func (b *EventBus) Subscribe(eventType types.EventType, handler types.EventHandlerFunc) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// Unsubscribe removes a handler for a specific event type
func (b *EventBus) Unsubscribe(eventType types.EventType, handler types.EventHandlerFunc) {
	b.lock.Lock()
	defer b.lock.Unlock()
	handlers := b.subscribers[eventType]
	for i, h := range handlers {
		if reflect.ValueOf(h).Pointer() == reflect.ValueOf(handler).Pointer() {
			b.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish sends an event to all subscribers
func (b *EventBus) Publish(event types.Event) {
	b.lock.RLock()
	handlers := b.subscribers[event.Type]
	defer b.lock.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}
