package ycq

import (
	. "gopkg.in/check.v1"
)

var _ = Suite(&GEventStoreSuite{})

type GEventStoreSuite struct {
	eventBus   *InternalEventBus
	eventStore *GetEventStore
}

func (s *GEventStoreSuite) SetUpTest(c *C) {
	s.eventBus = NewInternalEventBus()
	s.eventStore, _ = NewGetEventStore(s.eventBus, "")
}

func (s *GEventStoreSuite) TestNewEventStore(c *C) {
	es, err := NewGetEventStore(s.eventBus, "SomeURL")

	c.Assert(err, IsNil)
	c.Assert(es.eventBus, Equals, s.eventBus)
}

func (s *GEventStoreSuite) TestCanSetEventFactory(c *C) {
	eventFactory := NewDelegateEventFactory()
	s.eventStore.SetEventFactory(eventFactory)
	c.Assert(s.eventStore.eventFactory, Equals, eventFactory)
}
