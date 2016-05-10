package ycq

import (
	"net/http"

	"github.com/jetbasrawi/goes"
)

type GetEventStoreRepositoryClient interface {
	ReadStreamForwardAsync(string, *goes.StreamVersion, *goes.Take) <-chan *goes.AsyncResponse
	AppendToStream(string, *goes.StreamVersion, ...*goes.Event) (*goes.Response, error)
}

type FakeAsyncReader struct {
	eventResponses []*goes.EventResponse
	stream         string
	appended       []*goes.Event
}

func NewFakeAsyncClient() *FakeAsyncReader {
	fake := &FakeAsyncReader{}
	return fake
}

func (c *FakeAsyncReader) ReadStreamForwardAsync(stream string, version *goes.StreamVersion, take *goes.Take) <-chan *goes.AsyncResponse {

	c.stream = stream
	eventsChannel := make(chan *goes.AsyncResponse)

	go func() {
		for _, v := range c.eventResponses {
			eventsChannel <- &goes.AsyncResponse{v, nil, nil}
		}
		close(eventsChannel)
		return
	}()

	return eventsChannel

}

func (c *FakeAsyncReader) AppendToStream(stream string, expectedVersion *goes.StreamVersion, events ...*goes.Event) (*goes.Response, error) {
	c.appended = events
	c.stream = stream
	r := &goes.Response{
		StatusCode:    http.StatusCreated,
		StatusMessage: "201 Created",
	}
	return r, nil
}
