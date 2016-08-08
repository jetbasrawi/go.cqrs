// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"fmt"
)

//Dispatcher is the interface that should be implemented by command dispatcher
//
//The dispatcher is the mechanism through which commands are distributed to
//the appropriate command handler.
//
//Command handlers are registered with the dispatcher for a given command type.
//It is good practice in CQRS to have only one command handler for a given command.
//When a command is passed to the dispatcher it will look for the registered command
//handler and call that handler's Handle method passing the command message as an
//argument.
//
//Commands contained in a CommandMessage envelope are passed to the Dispatcher via
//the dispatch method.
type Dispatcher interface {
	Dispatch(CommandMessage) error
	RegisterHandler(CommandHandler, ...interface{}) error
}

//InMemoryDispatcher provides a lightweight and performant in process dispatcher
type InMemoryDispatcher struct {
	handlers map[string]CommandHandler
}

//NewInMemoryDispatcher constructs a new in memory dispatcher
func NewInMemoryDispatcher() *InMemoryDispatcher {
	b := &InMemoryDispatcher{
		handlers: make(map[string]CommandHandler),
	}
	return b
}

//Dispatch passes the CommandMessage on to all registered command handlers.
func (b *InMemoryDispatcher) Dispatch(command CommandMessage) error {
	if handler, ok := b.handlers[command.CommandType()]; ok {
		return handler.Handle(command)
	}
	return fmt.Errorf("The command bus does not have a handler for commands of type: %s", command.CommandType())
}

//RegisterHandler registers a command handler for the command types specified by the
//variadic commands parameter.
func (b *InMemoryDispatcher) RegisterHandler(handler CommandHandler, commands ...interface{}) error {
	for _, command := range commands {
		typeName := typeOf(command)
		if _, ok := b.handlers[typeName]; ok {
			return fmt.Errorf("Duplicate command handler registration with command bus for command of type: %s", typeName)
		}
		b.handlers[typeName] = handler
	}
	return nil
}
