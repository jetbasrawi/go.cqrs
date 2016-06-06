package ycq

import (
	"math/rand"

	. "gopkg.in/check.v1"
)

var _ = Suite(&EventSuite{})

type EventSuite struct {
}

type SomeEvent struct {
	Item  string
	Count int
}

type SomeOtherEvent struct {
	OrderID string
}

func NewTestEventMessage(id string) *EventDescriptor {
	ev := &SomeEvent{Item: yooid(), Count: rand.Intn(100)}
	return NewEventMessage(id, ev)
}

func (s *EventSuite) TestNewEventMessage(c *C) {
	id := yooid()
	ev := &SomeEvent{Item: "Some String", Count: 43}

	em := NewEventMessage(id, ev)

	c.Assert(em.id, Equals, id)
	c.Assert(em.event, Equals, ev)
	c.Assert(em.headers, NotNil)
}

func (s *EventSuite) TestShouldGetTypeOfEvent(c *C) {
	se := &SomeEvent{"Some String", 42}
	em := &EventDescriptor{event: se}
	c.Assert(em.EventType(), Equals, "SomeEvent")
}

//TODO: Do i need this still?
//func (s *EventSuite) TestShouldGetTypeOfAggregate(c *C) {
//em := &EventMessage{aggregate: &SomeAggregate{}}
//c.Assert(em.AggregateType(), Equals, "SomeAggregate")
//}

func (s *EventSuite) TestShouldGetHeaders(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(yooid(), ev)
	em.headers["a"] = "b"

	h := em.GetHeaders()

	c.Assert(h, DeepEquals, em.headers)
}

func (s *EventSuite) TestShouldGetEvent(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(yooid(), ev)
	got := em.Event()
	c.Assert(got, DeepEquals, em.event)
}

func (s *EventSuite) TestAddHeaderInt(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(yooid(), ev)

	em.SetHeader("a", 3)

	c.Assert(em.headers["a"], Equals, 3)
}

func (s *EventSuite) TestAddHeaderString(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(yooid(), ev)

	em.SetHeader("a", "abc")

	c.Assert(em.headers["a"], Equals, "abc")
}

func (s *EventSuite) TestAddHeaderStruct(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(yooid(), ev)

	em.SetHeader("a", ev)

	c.Assert(em.headers["a"], DeepEquals, ev)
}
