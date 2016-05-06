package ycq

import "github.com/jetbasrawi/yoono-uuid"

type AggregateRoot interface {
	AggregateID() uuid.UUID
	AggregateType() string
	Version() int
	IncrementVersion()
	Handle(CommandMessage) error
	Apply(events EventMessage)
	StoreEvent(EventMessage)
	GetChanges() []EventMessage
	ClearChanges()
}

type AggregateBase struct {
	id      uuid.UUID
	version int
	changes []EventMessage
}

func NewAggregateBase(id uuid.UUID) *AggregateBase {
	return &AggregateBase{
		id:      id,
		changes: []EventMessage{},
	}
}

func (a *AggregateBase) AggregateID() uuid.UUID {
	return a.id
}

func (a *AggregateBase) Version() int {
	return a.version
}

func (a *AggregateBase) IncrementVersion() {
	a.version++
}

func (a *AggregateBase) StoreEvent(event EventMessage) {
	a.changes = append(a.changes, event)
}

func (a *AggregateBase) GetChanges() []EventMessage {
	return a.changes
}

func (a *AggregateBase) ClearChanges() {
	a.changes = []EventMessage{}
}
