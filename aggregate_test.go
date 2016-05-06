package ycq

import (
	"github.com/jetbasrawi/yoono-uuid"
	. "gopkg.in/check.v1"
)

var _ = Suite(&AggregateBaseSuite{})

type AggregateBaseSuite struct{}

func (s *AggregateBaseSuite) TestNewAggregateBase(c *C) {
	id := yooid()

	agg := NewAggregateBase(id)

	c.Assert(agg, NotNil)
	c.Assert(agg.AggregateID(), Equals, id)
	c.Assert(agg.Version(), Equals, 0)
}

func (s *AggregateBaseSuite) TestIncrementVersion(c *C) {
	agg := NewAggregateBase(yooid())
	c.Assert(agg.Version(), Equals, 0)

	agg.IncrementVersion()

	c.Assert(agg.Version(), Equals, 1)
}

func (s *AggregateBaseSuite) TestStoreOneEvent(c *C) {
	ev := NewTestEventMessage(yooid())
	agg := NewSomeAggregate(ev.AggregateID())

	agg.StoreEvent(ev)

	c.Assert(agg.GetChanges(), DeepEquals, []EventMessage{ev})
}

func (s *AggregateBaseSuite) TestStoreMultipleEvents(c *C) {
	agg := NewAggregateBase(yooid())
	ev1 := NewTestEventMessage(agg.AggregateID())
	ev2 := NewTestEventMessage(agg.AggregateID())

	agg.StoreEvent(ev1)
	agg.StoreEvent(ev2)

	c.Assert(agg.GetChanges(), DeepEquals, []EventMessage{ev1, ev2})
}

func (s *AggregateBaseSuite) TestClearUncommittedEvents(c *C) {
	agg := NewAggregateBase(yooid())
	ev := NewTestEventMessage(agg.AggregateID())

	agg.StoreEvent(ev)

	c.Assert(agg.GetChanges(), DeepEquals, []EventMessage{ev})
	agg.ClearChanges()
	c.Assert(agg.GetChanges(), DeepEquals, []EventMessage{})
}

type SomeAggregate struct {
	*AggregateBase
	events []EventMessage
}

func NewSomeAggregate(id uuid.UUID) AggregateRoot {
	return &SomeAggregate{
		AggregateBase: NewAggregateBase(id),
	}
}

func (t *SomeAggregate) AggregateType() string {
	return "SomeAggregate"
}

func (t *SomeAggregate) Apply(event EventMessage) {
	t.events = append(t.events, event)
}

func (t *SomeAggregate) Handle(command CommandMessage) error {
	return nil
}

type SomeOtherAggregate struct {
	*AggregateBase
	events []EventMessage
}

func NewSomeOtherAggregate(id uuid.UUID) AggregateRoot {
	return &SomeOtherAggregate{
		AggregateBase: NewAggregateBase(id),
	}
}

func (t *SomeOtherAggregate) AggregateType() string {
	return "SomeOtherAggregate"
}

func (t *SomeOtherAggregate) Apply(event EventMessage) {
	t.events = append(t.events, event)
}

func (t *SomeOtherAggregate) Handle(command CommandMessage) error {
	return nil
}

type EmptyAggregate struct {
}
