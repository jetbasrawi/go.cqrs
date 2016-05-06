package ycq

import (
	"github.com/jetbasrawi/yoono-uuid"
)

type EventMessage interface {
	AggregateID() uuid.UUID
	GetHeaders() map[string]interface{}
	SetHeader(string, interface{})
	Event() interface{}
	EventType() string
}

type EventEnvelope struct {
	id        uuid.UUID
	aggregate AggregateRoot
	event     interface{}
	headers   map[string]interface{}
}

func NewEventMessage(aggregateID uuid.UUID, event interface{}) *EventEnvelope {
	return &EventEnvelope{
		id:      aggregateID,
		event:   event,
		headers: make(map[string]interface{}),
	}
}

func (c *EventEnvelope) EventType() string {
	return typeOf(c.event)
}

func (c *EventEnvelope) AggregateID() uuid.UUID {
	return c.id
}

func (c *EventEnvelope) GetHeaders() map[string]interface{} {
	return c.headers
}

func (c *EventEnvelope) Event() interface{} {
	return c.event
}

func (c *EventEnvelope) SetHeader(key string, value interface{}) {
	c.headers[key] = value
}
