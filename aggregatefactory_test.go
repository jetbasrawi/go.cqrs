// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"fmt"

	. "gopkg.in/check.v1"
)

var _ = Suite(&DelegateAggregateFactorySuite{})

type DelegateAggregateFactorySuite struct {
	factory *DelegateAggregateFactory
}

func (s *DelegateAggregateFactorySuite) SetUpTest(c *C) {
	s.factory = NewDelegateAggregateFactory()
}

func (s *DelegateAggregateFactorySuite) TestNewAggregateFactory(c *C) {
	factory := NewDelegateAggregateFactory()
	c.Assert(factory.delegates, NotNil)
}

func (s *DelegateAggregateFactorySuite) TestCanRegisterAggregateFactoryDelegate(c *C) {
	err := s.factory.RegisterDelegate(&SomeAggregate{},
		func(id string) AggregateRoot { return NewSomeAggregate(id) })

	c.Assert(err, IsNil)

	id := NewUUID()
	c.Assert(s.factory.delegates[typeOf(&SomeAggregate{})](id),
		DeepEquals,
		NewSomeAggregate(id))
}

func (s *DelegateAggregateFactorySuite) TestDuplicateAggregateFactoryRegistrationReturnsAnError(c *C) {
	err := s.factory.RegisterDelegate(&SomeAggregate{},
		func(id string) AggregateRoot { return NewSomeAggregate(id) })

	c.Assert(err, IsNil)

	err = s.factory.RegisterDelegate(&SomeAggregate{},
		func(id string) AggregateRoot { return NewSomeAggregate(id) })

	c.Assert(err, NotNil)
	c.Assert(err,
		DeepEquals,
		fmt.Errorf("Factory delegate already registered for type: \"%s\"",
			typeOf(&SomeAggregate{})))
}

func (s *DelegateAggregateFactorySuite) TestCanGetAggregateInstanceFromString(c *C) {
	_ = s.factory.RegisterDelegate(&SomeAggregate{},
		func(id string) AggregateRoot { return NewSomeAggregate(id) })

	id := NewUUID()
	ev := s.factory.GetAggregate(typeOf(&SomeAggregate{}), id)
	c.Assert(ev, DeepEquals, NewSomeAggregate(id))
}
