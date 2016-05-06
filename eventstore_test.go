package ycq

type StubEventStore struct {
	events []EventMessage
	loaded string
	stream string
}

func (m *StubEventStore) Save(stream string, events []EventMessage, expectedVersion *int, headers map[string]interface{}) error {
	m.events = append(m.events, events...)
	m.stream = stream
	return nil
}

func (m *StubEventStore) Load(stream string) ([]EventMessage, error) {
	m.loaded = stream
	return m.events, nil
}
