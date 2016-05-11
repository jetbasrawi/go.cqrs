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

type EventDescriptor struct {
	id      uuid.UUID
	event   interface{}
	headers map[string]interface{}
}

func NewEventMessage(aggregateID uuid.UUID, event interface{}) *EventDescriptor {
	return &EventDescriptor{
		id:      aggregateID,
		event:   event,
		headers: make(map[string]interface{}),
	}
}

func (c *EventDescriptor) EventType() string {
	return typeOf(c.event)
}

func (c *EventDescriptor) AggregateID() uuid.UUID {
	return c.id
}

func (c *EventDescriptor) GetHeaders() map[string]interface{} {
	return c.headers
}

func (c *EventDescriptor) Event() interface{} {
	return c.event
}

func (c *EventDescriptor) SetHeader(key string, value interface{}) {
	c.headers[key] = value
}
