package ycq

import (
	"time"

	"github.com/jetbasrawi/yoono-uuid"
)

// MemoryEventStore implements EventStore as an in memory structure.
type MemoryEventStore struct {
	eventBus         EventBus
	aggregateRecords map[string]*memoryAggregateRecord
}

// NewMemoryEventStore creates a new MemoryEventStore.
func NewMemoryEventStore(eventBus EventBus) *MemoryEventStore {
	s := &MemoryEventStore{
		eventBus:         eventBus,
		aggregateRecords: make(map[string]*memoryAggregateRecord),
	}
	return s
}

// Save appends all events in the event stream to the memory store.
func (s *MemoryEventStore) Save(stream string, events []EventMessage, expectedVersion *int, headers map[string]interface{}) error {
	if len(events) == 0 {
		return ErrNoEventsToAppend
	}

	for _, event := range events {
		r := &memoryEventRecord{
			eventType: event.EventType(),
			timestamp: time.Now(),
			event:     event,
		}

		if a, ok := s.aggregateRecords[event.AggregateID().String()]; ok {
			a.version++
			r.version = a.version
			a.events = append(a.events, r)
		} else {
			s.aggregateRecords[event.AggregateID().String()] = &memoryAggregateRecord{
				aggregateID: event.AggregateID(),
				version:     0,
				events:      []*memoryEventRecord{r},
			}
		}

		// Publish event on the bus.
		if s.eventBus != nil {
			s.eventBus.PublishEvent(event)
		}
	}

	return nil
}

// Load loads all events for the aggregate id from the memory store.
// Returns ErrNoEventsFound if no events can be found.
func (s *MemoryEventStore) Load(stream string) ([]EventMessage, error) {
	if a, ok := s.aggregateRecords[stream]; ok {
		events := make([]EventMessage, len(a.events))
		for i, r := range a.events {
			events[i] = r.event
		}
		return events, nil
	}

	return nil, ErrNoEventsFound
}

type memoryAggregateRecord struct {
	aggregateID uuid.UUID
	version     int
	events      []*memoryEventRecord
}

type memoryEventRecord struct {
	eventType string
	version   int
	timestamp time.Time
	event     EventMessage
}
