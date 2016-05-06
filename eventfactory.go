package ycq

import (
	"fmt"
)

type EventFactory interface {
	GetEvent(string) interface{}
}

type DelegateEventFactory struct {
	eventFactories map[string]func() interface{}
}

func NewDelegateEventFactory() *DelegateEventFactory {
	return &DelegateEventFactory{
		eventFactories: make(map[string]func() interface{}),
	}
}

func (t *DelegateEventFactory) RegisterDelegate(event interface{}, delegate func() interface{}) error {
	typeName := typeOf(event)
	if _, ok := t.eventFactories[typeName]; ok {
		return fmt.Errorf("Factory delegate already registered for type: \"%s\"", typeName)
	}
	t.eventFactories[typeName] = delegate
	return nil
}

func (t *DelegateEventFactory) GetEvent(typeName string) interface{} {
	if f, ok := t.eventFactories[typeName]; ok {
		return f()
	}
	return nil
}
