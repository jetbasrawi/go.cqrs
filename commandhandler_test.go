package ycq

import (
	"github.com/jetbasrawi/yoono-uuid"
	. "gopkg.in/check.v1"
)

var _ = Suite(&CommandHandlerSuite{})

type CommandHandlerSuite struct {
}

type TestCommandHandler struct {
	command CommandMessage
}

func (t *TestCommandHandler) Handle(command CommandMessage) error {
	t.command = command
	return nil
}

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
