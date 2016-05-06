package ycq

import (
	"fmt"

	"github.com/jetbasrawi/yoono-uuid"
	. "gopkg.in/check.v1"
)

var _ = Suite(&DelegateStreamNamerSuite{})

type DelegateStreamNamerSuite struct {
	namer *DelegateStreamNamer
}

func (s *DelegateStreamNamerSuite) SetUpTest(c *C) {
	s.namer = NewDelegateStreamNamer()
}

func (s *DelegateStreamNamerSuite) TestNewDelegateStreamNamer(c *C) {
	namer := NewDelegateStreamNamer()
	c.Assert(namer.delegates, NotNil)
}

func (s *DelegateStreamNamerSuite) TestCanRegisterStreamNameDelegate(c *C) {

	err := s.namer.RegisterDelegate(func(a string, id uuid.UUID) string { return id.String() + a },
		&SomeAggregate{},
	)
	c.Assert(err, IsNil)
	agg := NewSomeAggregate(yooid())
	f, _ := s.namer.delegates[agg.AggregateType()]
	stream := f(agg.AggregateType(), agg.AggregateID())
	c.Assert(stream, Equals, agg.AggregateID().String()+agg.AggregateType())
}

func (s *DelegateStreamNamerSuite) TestCanRegisterStreamNameDelegateWithMultipleAggregateRoots(c *C) {
	err := s.namer.RegisterDelegate(func(a string, id uuid.UUID) string { return id.String() + a },
		&SomeAggregate{},
		&SomeOtherAggregate{},
	)
	c.Assert(err, IsNil)

	ar1 := NewSomeAggregate(yooid())
	f, _ := s.namer.delegates[ar1.AggregateType()]
	stream := f(ar1.AggregateType(), ar1.AggregateID())
	c.Assert(stream, Equals, ar1.AggregateID().String()+ar1.AggregateType())

	ar2 := NewSomeOtherAggregate(yooid())
	f2, _ := s.namer.delegates[ar2.AggregateType()]
	stream2 := f2(ar2.AggregateType(), ar2.AggregateID())
	c.Assert(stream2, Equals, ar2.AggregateID().String()+ar2.AggregateType())
}

func (s *DelegateStreamNamerSuite) TestCanGetStreamName(c *C) {
	err := s.namer.RegisterDelegate(func(a string, id uuid.UUID) string { return id.String() + a },
		&SomeAggregate{},
	)
	c.Assert(err, IsNil)
	agg := NewSomeAggregate(yooid())
	stream, err := s.namer.GetStreamName(agg.AggregateType(), agg.AggregateID())
	c.Assert(err, IsNil)
	c.Assert(stream, Equals, agg.AggregateID().String()+agg.AggregateType())
}

func (s *DelegateStreamNamerSuite) TestGetStreamNameReturnsAnErrorIfNoDelegateRegisteredForAggregate(c *C) {
	err := s.namer.RegisterDelegate(func(a string, id uuid.UUID) string { return id.String() + a },
		&SomeAggregate{},
	)
	agg := NewSomeOtherAggregate(yooid())
	stream, err := s.namer.GetStreamName(agg.AggregateType(), agg.AggregateID())
	c.Assert(err, NotNil)
	c.Assert(stream, Equals, "")
	c.Assert(err, DeepEquals, fmt.Errorf("There is no stream name delegate for aggregate of type \"%s\"",
		agg.AggregateType()))

}

func (s *DelegateStreamNamerSuite) TestRegisteringAStreamNameDelegateMoreThanOnceReturnsAndError(c *C) {

	err := s.namer.RegisterDelegate(func(a string, id uuid.UUID) string { return id.String() + a },
		&SomeAggregate{},
	)
	c.Assert(err, IsNil)

	err = s.namer.RegisterDelegate(func(a string, id uuid.UUID) string { return id.String() + a },
		&SomeAggregate{},
	)
	c.Assert(err, DeepEquals,
		fmt.Errorf("The stream name delegate for \"%s\" is already registered with the stream namer.",
			NewSomeAggregate(yooid()).AggregateType()))
}
