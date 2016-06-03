package ycq

import (
	"fmt"
	"github.com/jetbasrawi/yoono-uuid"
)

//AggregateFactory returns aggregate instances of a specified type with the
//AggregateID set to the uuid provided.
//
//An aggregate factory is typically a dependency of the repository that will
//delegate instantiation of aggregate instances to the Aggregate factory.
type AggregateFactory interface {
	GetAggregate(string, uuid.UUID) AggregateRoot
}

//DelegateAggregateFactory is an implementation of the AggregateFactory interface
//that supports registration of delegate functions to perform aggregate instantiation.
type DelegateAggregateFactory struct {
	delegates map[string]func(uuid.UUID) AggregateRoot
}

//NewDelegateAggregateFactory contructs a new DelegateAggregateFactory
func NewDelegateAggregateFactory() *DelegateAggregateFactory {
	return &DelegateAggregateFactory{
		delegates: make(map[string]func(uuid.UUID) AggregateRoot),
	}
}

//RegisterDelegate is used to register a new funtion for instantiation of an
//aggregate instance.
//
// func(id uuid.UUID) AggregateRoot {return NewMyAggregateType(id)}
// func(id uuid.UUID) AggregateRoot { return &MyAggregateType{AggregateBase:NewAggregateBase(id)} }
func (t *DelegateAggregateFactory) RegisterDelegate(aggregate AggregateRoot, delegate func(uuid.UUID) AggregateRoot) error {
	typeName := typeOf(aggregate)
	if _, ok := t.delegates[typeName]; ok {
		return fmt.Errorf("Factory delegate already registered for type: \"%s\"", typeName)
	}
	t.delegates[typeName] = delegate
	return nil
}

//GetAggrete calls the delegate for the type specified and returns the result.
func (t *DelegateAggregateFactory) GetAggregate(typeName string, id uuid.UUID) AggregateRoot {
	if f, ok := t.delegates[typeName]; ok {
		return f(id)
	}
	return nil
}
