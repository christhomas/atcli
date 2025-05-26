package services

import (
	"atcli/src/types"
	"sync"
)

// EventBus manages event subscriptions and publishing
type EventBus struct {
	subscribers map[types.EventType][]types.EventHandlerFunc
	mutex       sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[types.EventType][]types.EventHandlerFunc),
	}
}

// Subscribe registers a handler for a specific event type
func (b *EventBus) Subscribe(eventType types.EventType, handler types.EventHandlerFunc) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// Publish sends an event to all subscribers
func (b *EventBus) Publish(event types.Event) {
	b.mutex.RLock()
	handlers := b.subscribers[event.Type]
	defer b.mutex.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}
