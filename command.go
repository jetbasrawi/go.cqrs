// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

// CommandMessage is the interface that a command message must implement.
type CommandMessage interface {

	// AggregateID returns the ID of the Aggregate that the command relates to
	AggregateID() string

	// Headers returns the key value collection of headers for the command.
	Headers() map[string]interface{}

	// SetHeader sets the value of the header specified by the key
	SetHeader(string, interface{})

	// Command returns the actual command which is the payload of the command message.
	Command() interface{}

	// CommandType returns a string descriptor of the command name
	CommandType() string
}

// CommandDescriptor is an implementation of the command message interface.
type CommandDescriptor struct {
	id      string
	command interface{}
	headers map[string]interface{}
}

// NewCommandMessage returns a new command descriptor
func NewCommandMessage(aggregateID string, command interface{}) *CommandDescriptor {
	return &CommandDescriptor{
		id:      aggregateID,
		command: command,
		headers: make(map[string]interface{}),
	}
}

// CommandType returns the command type name as a string
func (c *CommandDescriptor) CommandType() string {
	return typeOf(c.command)
}

// AggregateID returns the ID of the aggregate that the command relates to.
func (c *CommandDescriptor) AggregateID() string {
	return c.id
}

// Headers returns the collection of headers for the command.
func (c *CommandDescriptor) Headers() map[string]interface{} {
	return c.headers
}

// SetHeader sets the value of the header with the specified key
func (c *CommandDescriptor) SetHeader(key string, value interface{}) {
	c.headers[key] = value
}

// Command returns the actual command payload of the message.
func (c *CommandDescriptor) Command() interface{} {
	return c.command
}
