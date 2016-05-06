package ycq

import (
	"fmt"
	"github.com/jetbasrawi/yoono-uuid"
)

type AggregateFactory interface {
	GetAggregate(string, uuid.UUID) AggregateRoot
}

type DelegateAggregateFactory struct {
	delegates map[string]func(uuid.UUID) AggregateRoot
}

func NewDelegateAggregateFactory() *DelegateAggregateFactory {
	return &DelegateAggregateFactory{
		delegates: make(map[string]func(uuid.UUID) AggregateRoot),
	}
}

func (t *DelegateAggregateFactory) RegisterDelegate(aggregate AggregateRoot, delegate func(uuid.UUID) AggregateRoot) error {
	typeName := typeOf(aggregate)
	if _, ok := t.delegates[typeName]; ok {
		return fmt.Errorf("Factory delegate already registered for type: \"%s\"", typeName)
	}
	t.delegates[typeName] = delegate
	return nil
}

func (t *DelegateAggregateFactory) GetAggregate(typeName string, id uuid.UUID) AggregateRoot {
	if f, ok := t.delegates[typeName]; ok {
		return f(id)
	}
	return nil
}
