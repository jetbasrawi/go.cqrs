// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

// EventMessage is the interface that a command must implement.
type EventMessage interface {

	// AggregateID returns the ID of the Aggregate that the event relates to
	AggregateID() string

	// GetHeaders returns the key value collection of headers for the event.
	//
	// Headers are metadata about the event that do not form part of the
	// actual event but are still required to be persisted alongside the event.
	GetHeaders() map[string]interface{}

	// SetHeader sets the value of the header specified by the key
	SetHeader(string, interface{})

	// Returns the actual event which is the payload of the event message.
	Event() interface{}

	// EventType returns a string descriptor of the command name
	EventType() string
}

// EventDescriptor is an implementation of the event message interface.
type EventDescriptor struct {
	id      string
	event   interface{}
	headers map[string]interface{}
}

// NewEventMessage returns a new event descriptor
func NewEventMessage(aggregateID string, event interface{}) *EventDescriptor {
	return &EventDescriptor{
		id:      aggregateID,
		event:   event,
		headers: make(map[string]interface{}),
	}
}

// EventType returns the name of the event type as a string.
func (c *EventDescriptor) EventType() string {
	return typeOf(c.event)
}

// AggregateID returns the ID of the Aggregate that the event relates to.
func (c *EventDescriptor) AggregateID() string {
	return c.id
}

// GetHeaders returns the headers for the event.
func (c *EventDescriptor) GetHeaders() map[string]interface{} {
	return c.headers
}

// SetHeader sets the value of the header specified by the key
func (c *EventDescriptor) SetHeader(key string, value interface{}) {
	c.headers[key] = value
}

// Event the event payload of the event message
func (c *EventDescriptor) Event() interface{} {
	return c.event
}
