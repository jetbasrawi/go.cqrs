// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

// EventBus is the inteface that an event bus must implement.
type EventBus interface {
	PublishEvent(EventMessage)
	AddHandler(EventHandler, ...interface{})
}

// InternalEventBus provides a lightweight in process event bus
type InternalEventBus struct {
	eventHandlers map[string]map[EventHandler]struct{}
}

// NewInternalEventBus constructs a new InternalEventBus
func NewInternalEventBus() *InternalEventBus {
	b := &InternalEventBus{
		eventHandlers: make(map[string]map[EventHandler]struct{}),
	}
	return b
}

// PublishEvent publishes events to all registered event handlers
func (b *InternalEventBus) PublishEvent(event EventMessage) {
	if handlers, ok := b.eventHandlers[event.EventType()]; ok {
		for handler := range handlers {
			handler.Handle(event)
		}
	}
}

// AddHandler registers an event handler for all of the events specified in the
// variadic events parameter.
func (b *InternalEventBus) AddHandler(handler EventHandler, events ...interface{}) {

	for _, event := range events {
		typeName := typeOf(event)

		// There can be multiple handlers for any event.
		// Here we check that a map is initialized to hold these handlers
		// for a given type. If not we create one.
		if _, ok := b.eventHandlers[typeName]; !ok {
			b.eventHandlers[typeName] = make(map[EventHandler]struct{})
		}

		// Add this handler to the collection of handlers for the type.
		b.eventHandlers[typeName][handler] = struct{}{}
	}
}
