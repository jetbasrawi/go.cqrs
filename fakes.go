package ycq

import (
	"net/http"

	"github.com/jetbasrawi/goes"
)

//type GetEventStoreRepositoryClient interface {
////ReadStreamForwardAsync(string, *goes.StreamVersion, *goes.Take) <-chan struct {
///[>goes.EventResponse
///[>goes.Response
////error
////}
//AppendToStream(string, *goes.StreamVersion, ...*goes.Event) (*goes.Response, error)
//}

type FakeAsyncReader struct {
	eventResponses []*goes.EventResponse
	stream         string
	appended       []*goes.Event
}

func NewFakeAsyncClient() *FakeAsyncReader {
	fake := &FakeAsyncReader{}
	return fake
}

func (c *FakeAsyncReader) ReadStreamForwardAsync(stream string, version *goes.StreamVersion, take *goes.Take) <-chan struct {
	*goes.EventResponse
	*goes.Response
	error
} {

	eventsChannel := make(chan struct {
		*goes.EventResponse
		*goes.Response
		error
	})

	go func() {
		for _, v := range c.eventResponses {
			eventsChannel <- struct {
				*goes.EventResponse
				*goes.Response
				error
			}{v, nil, nil}
		}
		close(eventsChannel)
		return
	}()

	return eventsChannel

}

func (c *FakeAsyncReader) AppendToStream(stream string, expectedVersion *goes.StreamVersion, events ...*goes.Event) (*goes.Response, error) {
	c.appended = events
	r := &goes.Response{
		StatusCode:    http.StatusCreated,
		StatusMessage: "201 Created",
	}
	return r, nil
}
