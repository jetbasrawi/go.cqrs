package ycq

import (
	"fmt"
)

type Dispatcher interface {
	Dispatch(CommandMessage) error
	RegisterHandler(CommandHandler, ...interface{}) error
}

type InMemoryDispatcher struct {
	handlers map[string]CommandHandler
}

func NewInMemoryDispatcher() *InMemoryDispatcher {
	b := &InMemoryDispatcher{
		handlers: make(map[string]CommandHandler),
	}
	return b
}

func (b *InMemoryDispatcher) Dispatch(command CommandMessage) error {
	if handler, ok := b.handlers[command.CommandType()]; ok {
		return handler.Handle(command)
	}
	return fmt.Errorf("The command bus does not have a handler for commands of type: %s", command.CommandType())
}

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
