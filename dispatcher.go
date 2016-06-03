package ycq

import (
	"fmt"
)

//Dispatcher is the interface that should be implemented by command dispatchers
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

//Dispatch passes the EventMessage on to all registered command handlers.
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
			return fmt.Errorf("Duplicate command handler registration with command bus for command of type: %d", typeName)
		}
		b.handlers[typeName] = handler
	}
	return nil
}
