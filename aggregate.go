package ycq

import "github.com/jetbasrawi/yoono-uuid"

//AggregateRoot is the interface that all aggregates should implement
type AggregateRoot interface {
	AggregateID() uuid.UUID
	Version() int
	IncrementVersion()
	Apply(events EventMessage, isNew bool)
	TrackChange(EventMessage)
	GetChanges() []EventMessage
	ClearChanges()
}

//AggregateBase is a type that can be embedded in an AggregateRoot
//implementation to handle common aggragate behaviour
//
//All required methods to implement an aggregate are here, to implement the
//Aggregate root interface your aggregate will need to implement the Apply
//method that will contain behaviour specific to your aggregate.
type AggregateBase struct {
	id      uuid.UUID
	version int
	changes []EventMessage
}

//NewAggregateBase contructs a new AggregateBase.
func NewAggregateBase(id uuid.UUID) *AggregateBase {
	return &AggregateBase{
		id:      id,
		changes: []EventMessage{},
	}
}

//AggregateID returns the AggregateID
func (a *AggregateBase) AggregateID() uuid.UUID {
	return a.id
}

//Version returns the version of the aggregate.
func (a *AggregateBase) Version() int {
	return a.version
}

//IncrementVersion increments the aggregate version number
func (a *AggregateBase) IncrementVersion() {
	a.version++
}

//TrackChange stores the EventMessage in the changes collection.
//
//Changes are new, unpersisted events that have been applied to the aggregate.
func (a *AggregateBase) TrackChange(event EventMessage) {
	a.changes = append(a.changes, event)
}

//GetChanges returns the collection of new unpersisted events that have
//been applied to the aggregate.
func (a *AggregateBase) GetChanges() []EventMessage {
	return a.changes
}

//ClearChanges removes all unpersisted events from the aggregate.
func (a *AggregateBase) ClearChanges() {
	a.changes = []EventMessage{}
}
