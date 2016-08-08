// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"fmt"

	. "gopkg.in/check.v1"
)

var _ = Suite(&InternalCommandBusSuite{})

type InternalCommandBusSuite struct {
	bus         *InMemoryDispatcher
	stubhandler *TestCommandHandler
}

func (s *InternalCommandBusSuite) SetUpTest(c *C) {
	s.bus = NewInMemoryDispatcher()
	s.stubhandler = &TestCommandHandler{}
}

func (s *InternalCommandBusSuite) TestNewInternalCommandBus(c *C) {
	bus := NewInMemoryDispatcher()
	c.Assert(bus, NotNil)
}

func (s *InternalCommandBusSuite) TestShouldHandleCommand(c *C) {
	err := s.bus.RegisterHandler(s.stubhandler, &SomeCommand{})
	c.Assert(err, IsNil)
	cmd := NewSomeCommandMessage(NewUUID())

	err = s.bus.Dispatch(cmd)

	c.Assert(err, IsNil)
	c.Assert(s.stubhandler.command, Equals, cmd)
}

func (s *InternalCommandBusSuite) TestShouldReturnErrorIfNoHandlerRegisteredForCommand(c *C) {
	cmd := NewSomeCommandMessage(NewUUID())

	err := s.bus.Dispatch(cmd)

	c.Assert(err, DeepEquals, fmt.Errorf("The command bus does not have a handler for commands of type: %s", cmd.CommandType()))
	c.Assert(s.stubhandler.command, IsNil)
}

func (s *InternalCommandBusSuite) TestDuplicateHandlerRegistrationReturnsAnError(c *C) {
	err := s.bus.RegisterHandler(s.stubhandler, &SomeCommand{}, &SomeCommand{})
	c.Assert(err, DeepEquals, fmt.Errorf("Duplicate command handler registration with command bus for command of type: %s",
		typeOf(&SomeCommand{"", 0})))
}

func (s *InternalCommandBusSuite) TestCanRegisterMultipleCommandsForTheSameHandler(c *C) {

	err := s.bus.RegisterHandler(s.stubhandler, &SomeCommand{}, &SomeOtherCommand{})
	c.Assert(err, IsNil)

}
