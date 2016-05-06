package ycq

import "github.com/jetbasrawi/yoono-uuid"

type MockRepository struct {
	aggregates map[uuid.UUID]AggregateRoot
}

func (m *MockRepository) Load(aggregateType string, id uuid.UUID) (AggregateRoot, error) {
	return m.aggregates[id], nil
}

func (m *MockRepository) Save(aggregate AggregateRoot) error {
	m.aggregates[aggregate.AggregateID()] = aggregate
	return nil
}
