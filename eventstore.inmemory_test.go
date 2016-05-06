package ycq

import (
	. "gopkg.in/check.v1"
)

var _ = Suite(&MemoryEventStoreSuite{})

type MemoryEventStoreSuite struct {
	store *MemoryEventStore
	bus   *MockEventBus
}

func (s *MemoryEventStoreSuite) SetUpTest(c *C) {
	s.bus = &MockEventBus{
		events: make([]EventMessage, 0),
	}
	s.store = NewMemoryEventStore(s.bus)
}

func (s *MemoryEventStoreSuite) Test_NewMemoryEventStore(c *C) {
	bus := &MockEventBus{
		events: make([]EventMessage, 0),
	}
	store := NewMemoryEventStore(bus)
	c.Assert(store, NotNil)
}

func (s *MemoryEventStoreSuite) Test_NoEvents(c *C) {
	err := s.store.Save("", []EventMessage{}, nil, nil)
	c.Assert(err, Equals, ErrNoEventsToAppend)
}

//func (s *MemoryEventStoreSuite) Test_OneEvent(c *C) {
//event1 := &TestEvent{yooid(), "event1"}
//err := s.store.Save("", []IEventMessage{event1}, nil, nil)
//c.Assert(err, IsNil)
//events, err := s.store.Load(event1.TestID.String())
//c.Assert(err, IsNil)
//c.Assert(events, HasLen, 1)
//c.Assert(events[0], DeepEquals, event1)
//c.Assert(s.bus.events, DeepEquals, events)
//}

//func (s *MemoryEventStoreSuite) Test_TwoEvents(c *C) {
//event1 := &TestEvent{yooid(), "event1"}
//event2 := &TestEvent{event1.TestID, "event2"}
//err := s.store.Save("", []IEventMessage{event1, event2}, nil, nil)
//c.Assert(err, IsNil)
//events, err := s.store.Load(event1.TestID.String())
//c.Assert(err, IsNil)
//c.Assert(events, HasLen, 2)
//c.Assert(events[0], DeepEquals, event1)
//c.Assert(events[1], DeepEquals, event2)
//c.Assert(s.bus.events, DeepEquals, events)
//}

//func (s *MemoryEventStoreSuite) Test_DifferentAggregates(c *C) {
//event1 := &TestEvent{yooid(), "event1"}
//event2 := &TestEvent{yooid(), "event2"}
//err := s.store.Save("", []IEventMessage{event1, event2}, nil, nil)
//c.Assert(err, IsNil)
//events, err := s.store.Load(event1.TestID.String())
//c.Assert(err, IsNil)
//c.Assert(events, HasLen, 1)
//c.Assert(events[0], DeepEquals, event1)
//events, err = s.store.Load(event2.TestID.String())
//c.Assert(err, IsNil)
//c.Assert(events, HasLen, 1)
//c.Assert(events[0], DeepEquals, event2)
//c.Assert(s.bus.events, DeepEquals, []IEventMessage{event1, event2})
//}

//func (s *MemoryEventStoreSuite) Test_LoadNoEvents(c *C) {
//events, err := s.store.Load(yooid().String())
//c.Assert(err, ErrorMatches, "could not find events")
//c.Assert(events, DeepEquals, []IEventMessage(nil))
//}
