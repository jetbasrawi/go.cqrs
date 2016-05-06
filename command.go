package ycq

import (
	"github.com/jetbasrawi/yoono-uuid"
)

type CommandMessage interface {
	AggregateID() uuid.UUID
	Headers() map[string]interface{}
	SetHeader(string, interface{})
	Command() interface{}
	CommandType() string
}

type CmdMessage struct {
	id      uuid.UUID
	command interface{}
	headers map[string]interface{}
}

func NewCommandMessage(aggregateID uuid.UUID, command interface{}) *CmdMessage {
	return &CmdMessage{
		id:      aggregateID,
		command: command,
		headers: make(map[string]interface{}),
	}
}

func (c *CmdMessage) CommandType() string {
	return typeOf(c.command)
}

func (c *CmdMessage) AggregateID() uuid.UUID {
	return c.id
}

func (c *CmdMessage) Headers() map[string]interface{} {
	return c.headers
}

func (c *CmdMessage) Command() interface{} {
	return c.command
}

func (c *CmdMessage) SetHeader(key string, value interface{}) {
	c.headers[key] = value
}
