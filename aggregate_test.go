// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import . "gopkg.in/check.v1"

var _ = Suite(&AggregateBaseSuite{})

type AggregateBaseSuite struct{}

func (s *AggregateBaseSuite) TestNewAggregateBase(c *C) {
	id := NewUUID()

	agg := NewAggregateBase(id)

	c.Assert(agg, NotNil)
	c.Assert(agg.AggregateID(), Equals, id)
	c.Assert(agg.OriginalVersion(), Equals, -1)
	c.Assert(agg.CurrentVersion(), Equals, -1)
}

func (s *AggregateBaseSuite) TestIncrementVersion(c *C) {
	agg := NewAggregateBase(NewUUID())
	c.Assert(agg.CurrentVersion(), Equals, -1)

	agg.IncrementVersion()
	c.Assert(agg.CurrentVersion(), Equals, 0)
}

func (s *AggregateBaseSuite) TestTrackOneChange(c *C) {
	ev := NewTestEventMessage(NewUUID())
	agg := NewSomeAggregate(ev.AggregateID())

	agg.TrackChange(ev)

	c.Assert(agg.GetChanges(), DeepEquals, []EventMessage{ev})
}

func (s *AggregateBaseSuite) TestTrackMultipleChanges(c *C) {
	agg := NewAggregateBase(NewUUID())
	ev1 := NewTestEventMessage(agg.AggregateID())
	ev2 := NewTestEventMessage(agg.AggregateID())

	agg.TrackChange(ev1)
	agg.TrackChange(ev2)

	c.Assert(agg.GetChanges(), DeepEquals, []EventMessage{ev1, ev2})
}

func (s *AggregateBaseSuite) TestClearChanges(c *C) {
	agg := NewAggregateBase(NewUUID())
	ev := NewTestEventMessage(agg.AggregateID())

	agg.TrackChange(ev)

	c.Assert(agg.GetChanges(), DeepEquals, []EventMessage{ev})
	agg.ClearChanges()
	c.Assert(agg.GetChanges(), DeepEquals, []EventMessage{})
}

type SomeAggregate struct {
	*AggregateBase
	events []EventMessage
}

func NewSomeAggregate(id string) AggregateRoot {
	return &SomeAggregate{
		AggregateBase: NewAggregateBase(id),
	}
}

func (t *SomeAggregate) Apply(event EventMessage, isNew bool) {
	t.events = append(t.events, event)
}

func (t *SomeAggregate) Handle(command CommandMessage) error {
	return nil
}

type SomeOtherAggregate struct {
	*AggregateBase
	changes []EventMessage
}

func NewSomeOtherAggregate(id string) AggregateRoot {
	return &SomeOtherAggregate{
		AggregateBase: NewAggregateBase(id),
	}
}

//TODO: No tests for isNew
func (t *SomeOtherAggregate) Apply(event EventMessage, isNew bool) {
	t.changes = append(t.changes, event)
}

func (t *SomeOtherAggregate) Handle(command CommandMessage) error {
	return nil
}

type EmptyAggregate struct {
}
