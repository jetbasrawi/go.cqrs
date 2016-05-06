package ycq

type TraceEventStore struct {
	eventStore EventStore
	tracing    bool
	trace      []EventMessage
}

func NewTraceEventStore(eventStore EventStore) *TraceEventStore {
	s := &TraceEventStore{
		eventStore: eventStore,
		trace:      make([]EventMessage, 0),
	}
	return s
}

func (s *TraceEventStore) Save(stream string, events []EventMessage, expectedVersion *int, headers map[string]interface{}) error {
	if s.tracing {
		s.trace = append(s.trace, events...)
	}

	if s.eventStore != nil {
		return s.eventStore.Save(stream, events, nil, nil)
	}

	return nil
}

func (s *TraceEventStore) Load(stream string) ([]EventMessage, error) {
	if s.eventStore != nil {
		return s.eventStore.Load(stream)
	}

	return nil, ErrNoEventStoreDefined
}

// StartTracing starts the tracing of events.
func (s *TraceEventStore) StartTracing() {
	s.tracing = true
}

// StopTracing stops the tracing of events.
func (s *TraceEventStore) StopTracing() {
	s.tracing = false
}

// GetTrace returns the events that happened during the tracing.
func (s *TraceEventStore) GetTrace() []EventMessage {
	return s.trace
}

// ResetTrace resets the trace.
func (s *TraceEventStore) ResetTrace() {
	s.trace = make([]EventMessage, 0)
}
