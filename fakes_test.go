package ycq

import (
	//"net/http"

	//"github.com/jetbasrawi/goes"
	. "gopkg.in/check.v1"
)

var _ = Suite(&FakeSuite{})

type FakeSuite struct{}

func (s *FakeSuite) TestAsyncFakeReadForwardAll(c *C) {

	//stream := "some-stream"
	//es := goes.CreateTestEvents(10, stream, "some-server", "MyEventType")
	//ers := goes.CreateTestEventResponses(es, nil)
	//fake := NewFakeAsyncClient()
	//fake.eventResponses = ers
	//fake.stream = stream

	//eventsChannel := fake.ReadStreamForwardAsync(stream, nil, nil)
	//count := 0
	//for {
	//select {
	//case ev, open := <-eventsChannel:
	//if !open {
	//c.Assert(count, Equals, len(es))
	//return
	//}

	//c.Assert(ev.error, IsNil)
	//c.Assert(ev.EventResponse, Equals, es[count])
	//count++
	//}
	//}
}

func (s *FakeSuite) TestAppendToStream(c *C) {

	//stream := "some-stream"
	//es := goes.CreateTestEvents(10, stream, "some-server", "MyEventType")
	//ers := goes.CreateTestEventResponses(es, nil)
	//fake := NewFakeAsyncClient()

	//resp, err := fake.AppendToStream(stream, nil, es...)
	//c.Assert(err, IsNil)
	//c.Assert(resp.StatusCode, Equals, http.StatusCreated)
	//c.Assert(resp.StatusMessage, Equals, "201 Created")

	//c.Assert(fake.appended, Equals, es)

}
