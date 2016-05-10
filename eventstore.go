package ycq

import (
	"errors"
	//"github.com/jetbasrawi/yoono-uuid"
)

// Error returned when no events are available to append.
var ErrNoEventsToAppend = errors.New("no events to append")

// Error returned when no events are found.
var ErrNoEventsFound = errors.New("could not find events")

// Error returned if no event store has been defined.
var ErrNoEventStoreDefined = errors.New("no event store defined")

// EventStore is an interface for an event sourcing event store.
type EventStore interface {
	// Save appends all events in the event stream to the store.
	// First argument is the stream name
	// Second argument is the events to be saved
	// Third argument is the expected version
	// Fourth argument is the headers
	Save(string, []EventMessage, *int, map[string]interface{}) error

	// Load loads all events for the aggregate identified by the stream name from the store.
	Load(string) ([]EventMessage, error)
}
