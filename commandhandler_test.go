package ycq

import (
	"fmt"
	"github.com/jetbasrawi/yoono-uuid"
	. "gopkg.in/check.v1"
)

var _ = Suite(&AggregateCommandHandlerSuite{})

type AggregateCommandHandlerSuite struct {
	repo    *MockRepository
	handler *AggregateCommandHandler
}

func (s *AggregateCommandHandlerSuite) SetUpTest(c *C) {
	s.repo = &MockRepository{
		aggregates: make(map[uuid.UUID]AggregateRoot),
	}
	s.handler, _ = NewAggregateCommandHandler(s.repo)
}

func (s *AggregateCommandHandlerSuite) TestNewDispatcher(c *C) {
	repo := &MockRepository{
		aggregates: make(map[uuid.UUID]AggregateRoot),
	}
	handler, err := NewAggregateCommandHandler(repo)
	c.Assert(handler, NotNil)
	c.Assert(err, IsNil)
}

func (s *AggregateCommandHandlerSuite) TestDispatcherReturnsErrorIfNoRepositoryInjected(c *C) {
	handler, err := NewAggregateCommandHandler(nil)
	c.Assert(handler, IsNil)
	c.Assert(err, ErrorMatches, "The command dispatcher requires a repository.")
}

func (s *AggregateCommandHandlerSuite) TestDispatcherShouldDispatchCommandToAggregate(c *C) {
	agg := &TestDispatcherAggregate{
		AggregateBase: NewAggregateBase(yooid()),
	}
	s.repo.aggregates[agg.AggregateID()] = agg
	s.handler.RegisterCommands(agg, &SomeCommand{})
	cmd := NewSomeCommandMessage(agg.AggregateID())

	err := s.handler.Handle(cmd)

	c.Assert(agg.command, Equals, cmd)
	c.Assert(err, IsNil)
}

func (s *AggregateCommandHandlerSuite) TestDispatcherReturnsErrorsFromTheHandlers(c *C) {
	agg := &TestDispatcherAggregate{
		AggregateBase: NewAggregateBase(yooid()),
	}
	s.repo.aggregates[agg.AggregateID()] = agg
	s.handler.RegisterCommands(agg, &ErrorCommand{})
	cmd := NewSomeCommandMessage(agg.AggregateID())
	body := &ErrorCommand{Message: "Unable to process command."}
	cmd.command = body

	err := s.handler.Handle(cmd)

	c.Assert(err, ErrorMatches, body.Message)
}

func (s *AggregateCommandHandlerSuite) TestDispatcherReturnsErrorIfNoHandlerRegisteredForACommand(c *C) {
	cmd := NewSomeCommandMessage(yooid())
	err := s.handler.Handle(cmd)
	c.Assert(err, DeepEquals, fmt.Errorf("The dispatcher has no handler registered for commands of type: \"%s\"", cmd.CommandType()))
}

func (s *AggregateCommandHandlerSuite) TestDispatcherReturnsErrorIfCommandAlreadyRegistered(c *C) {
	err := s.handler.RegisterCommands(&SomeAggregate{}, &SomeCommand{}, &ErrorCommand{}, &SomeCommand{})
	c.Assert(err, DeepEquals, fmt.Errorf("The command \"%s\" is already registered with the dispatcher.", typeOf(&SomeCommand{})))
}

///////////////////////////////////////////////////////////////////////////////
// Fakes

type TestDispatcherAggregate struct {
	*AggregateBase
	command CommandMessage
}

func (t *TestDispatcherAggregate) AggregateType() string {
	return typeOf(t)
}

func (t *TestDispatcherAggregate) Handle(command CommandMessage) error {
	t.command = command

	switch cmd := command.Command().(type) {
	case *SomeCommand:
		t.TrackChange(NewEventMessage(command.AggregateID(), &SomeEvent{cmd.Item, cmd.Count}))
		return nil
	case *ErrorCommand:
		return fmt.Errorf(cmd.Message)
	}

	return fmt.Errorf("The command %s was not handled", command.CommandType())
}

func (t *TestDispatcherAggregate) Apply(event EventMessage) {}

type TestCommandHandler struct {
	command CommandMessage
}

func (t *TestCommandHandler) Handle(command CommandMessage) error {
	t.command = command
	return nil
}
