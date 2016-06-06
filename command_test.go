package ycq

import (
	"math/rand"

	. "gopkg.in/check.v1"
)

var _ = Suite(&CommandSuite{})

type CommandSuite struct{}

type SomeCommand struct {
	Item  string
	Count int
}

func NewSomeCommandMessage(id string) *CommandDescriptor {
	ev := &SomeCommand{Item: yooid(), Count: rand.Intn(100)}
	return NewCommandMessage(id, ev)
}

type SomeOtherCommand struct {
	OrderID string
}

func NewSomeOtherCommandMessage(id string) *CommandDescriptor {
	ev := &SomeOtherCommand{id}
	return NewCommandMessage(id, ev)
}

type ErrorCommand struct {
	Message string
}

func (s *CommandSuite) TestNewCommandMessage(c *C) {
	id := yooid()
	cmd := &SomeCommand{Item: "Some String", Count: 43}

	cm := NewCommandMessage(id, cmd)

	c.Assert(cm.id, Equals, id)
	c.Assert(cm.command, Equals, cmd)
	c.Assert(cm.headers, NotNil)
}

func (s *CommandSuite) TestShouldGetTypeOfCommand(c *C) {
	sc := &SomeCommand{"Some String", 42}
	cm := &CommandDescriptor{command: sc}

	typeString := cm.CommandType()

	c.Assert(typeString, Equals, "SomeCommand")
}

func (s *CommandSuite) TestShouldGetHeaders(c *C) {
	cmd := &SomeCommand{"Some data", 456}
	cm := NewCommandMessage(yooid(), cmd)
	cm.headers["a"] = "b"

	h := cm.Headers()

	c.Assert(h, DeepEquals, cm.headers)
}

func (s *CommandSuite) TestShouldGetCommand(c *C) {
	cmd := &SomeCommand{"Some data", 456}
	cm := NewCommandMessage(yooid(), cmd)

	got := cm.Command()

	c.Assert(got, DeepEquals, cm.command)
}

func (s *CommandSuite) TestAddHeaderInt(c *C) {
	cmd := &SomeCommand{"Some data", 456}
	cm := NewCommandMessage(yooid(), cmd)

	cm.SetHeader("a", 3)

	c.Assert(cm.headers["a"], Equals, 3)
}

func (s *CommandSuite) TestAddHeaderString(c *C) {
	cmd := &SomeCommand{"Some data", 456}
	cm := NewCommandMessage(yooid(), cmd)

	cm.SetHeader("a", "abc")

	c.Assert(cm.headers["a"], Equals, "abc")
}

func (s *CommandSuite) TestAddHeaderStruct(c *C) {
	cmd := &SomeCommand{"Some data", 456}
	cm := NewCommandMessage(yooid(), cmd)

	cm.SetHeader("a", cmd)

	c.Assert(cm.headers["a"], DeepEquals, cmd)
}
