package ycq

import (
	"fmt"
	"net/http"

	"github.com/jetbasrawi/goes"
	"github.com/jetbasrawi/yoono-uuid"
	. "gopkg.in/check.v1"
)

var _ = Suite(&ComDomRepoSuite{})

type ComDomRepoSuite struct {
	store          *FakeAsyncReader
	repo           *CommonDomainRepository
	someEvent      *SomeEvent
	someOtherEvent *SomeOtherEvent
}

func (s *ComDomRepoSuite) SetUpTest(c *C) {
	store := NewFakeAsyncClient()
	stream := "astream"
	server := "http://localhost:2113"
	s.someEvent = &SomeEvent{Item: "Some Item", Count: 42}
	goesEvent := goes.CreateTestEventFromData(stream, server, 0, s.someEvent, nil)
	s.someOtherEvent = &SomeOtherEvent{yooid()}
	goesOtherEvent := goes.CreateTestEventFromData(stream, server, 1, s.someOtherEvent, nil)
	es := []*goes.Event{goesEvent, goesOtherEvent}
	ers := goes.CreateTestEventResponses(es, nil)
	store.eventResponses = ers
	store.stream = stream
	s.store = store

	eventBus := NewInternalEventBus()

	s.repo, _ = NewCommonDomainRepository(s.store, eventBus)

	aggregateFactory := NewDelegateAggregateFactory()
	aggregateFactory.RegisterDelegate(&SomeAggregate{},
		func(id uuid.UUID) AggregateRoot { return NewSomeAggregate(id) })
	s.repo.aggregateFactory = aggregateFactory

	streamNameDelegate := NewDelegateStreamNamer()
	streamNameDelegate.RegisterDelegate(func(at string, id uuid.UUID) string { return at + id.String() },
		&SomeAggregate{},
		&SomeOtherAggregate{},
		&StubAggregate{})
	s.repo.streamNameDelegate = streamNameDelegate

	eventFactory := NewDelegateEventFactory()
	eventFactory.RegisterDelegate(&SomeEvent{},
		func() interface{} { return &SomeEvent{} })
	eventFactory.RegisterDelegate(&SomeOtherEvent{},
		func() interface{} { return &SomeOtherEvent{} })
	s.repo.SetEventFactory(eventFactory)
}

func (s *ComDomRepoSuite) TestCanConstructNewRepository(c *C) {
	store, _ := goes.NewClient(nil, "")
	eventBus := NewInternalEventBus()

	repo, err := NewCommonDomainRepository(store, eventBus)

	c.Assert(repo, NotNil)
	c.Assert(err, IsNil)
	c.Assert(repo.aggregateFactory, IsNil)
	c.Assert(repo.streamNameDelegate, IsNil)
	c.Assert(repo.eventBus, NotNil)
}

func (s *ComDomRepoSuite) TestCreatingNewRepositoryWithNilEventStoreReturnsAnError(c *C) {
	eventBus := NewInternalEventBus()
	repo, err := NewCommonDomainRepository(nil, eventBus)

	c.Assert(repo, IsNil)
	c.Assert(err, DeepEquals, fmt.Errorf("Nil Eventstore injected into repository."))
}

func (s *ComDomRepoSuite) TestCreatingNewRepositoryWithNilEventBusReturnsAnError(c *C) {
	store, _ := goes.NewClient(nil, "")
	repo, err := NewCommonDomainRepository(store, nil)

	c.Assert(repo, IsNil)
	c.Assert(err, DeepEquals, fmt.Errorf("Nil EventBus injected into repository."))
}

func (s *ComDomRepoSuite) TestRepositoryCanLoadAnAggregate(c *C) {
	s.store.eventResponses = []*goes.EventResponse{}
	id := yooid()

	agg, err := s.repo.Load(typeOf(&SomeAggregate{}), id)

	c.Assert(err, IsNil)
	c.Assert(agg.AggregateID(), Equals, id)
	c.Assert(typeOf(agg), Equals, typeOf(NewSomeAggregate(id)))
	c.Assert(agg.Version(), Equals, 0)
}

func (s *ComDomRepoSuite) TestRepositoryCanLoadAggregateWithEvents(c *C) {
	id := yooid()
	aggregateFactory := NewDelegateAggregateFactory()
	aggregateFactory.RegisterDelegate(&StubAggregate{},
		func(id uuid.UUID) AggregateRoot { return NewStubAggregate(id) })
	s.repo.SetAggregateFactory(aggregateFactory)

	got, err := s.repo.Load(typeOf(&StubAggregate{}), id)

	c.Assert(err, IsNil)
	c.Assert(got.AggregateID(), Equals, id)

	events := got.(*StubAggregate).events
	c.Assert(events[0].Event(), DeepEquals, s.someEvent)
	c.Assert(events[1].Event(), DeepEquals, s.someOtherEvent)
}

func (s *ComDomRepoSuite) TestRepositoryIncrementsAggregateVersionForEachEvent(c *C) {
	stream := "astream"
	server := "http://localhost:2113"
	ev1 := goes.CreateTestEventFromData(stream, server, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	ev2 := goes.CreateTestEventFromData(stream, server, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	ev3 := goes.CreateTestEventFromData(stream, server, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	es := []*goes.Event{ev1, ev2, ev3}
	s.store.eventResponses = goes.CreateTestEventResponses(es, nil)
	id := yooid()

	got, _ := s.repo.Load(typeOf(&SomeAggregate{}), id)

	c.Assert(got.Version(), Equals, 3)
}

func (s *ComDomRepoSuite) TestSaveAggregateWithUncommittedChanges(c *C) {
	s.store.eventResponses = []*goes.EventResponse{}
	id := yooid()
	agg := NewSomeAggregate(id)
	ev := &SomeEvent{Item: "Some string", Count: 4353}
	em := NewEventMessage(id, ev)
	agg.TrackChange(em)

	err := s.repo.Save(agg)

	c.Assert(err, IsNil)
	c.Assert(s.store.appended[0].Data, DeepEquals, ev)
}

func (s *ComDomRepoSuite) TestCanRegisterAggregateFactory(c *C) {
	aggregateFactory := NewDelegateAggregateFactory()

	s.repo.SetAggregateFactory(aggregateFactory)

	c.Assert(s.repo.aggregateFactory, Equals, aggregateFactory)
}

func (s *ComDomRepoSuite) TestNoAggregateFactoryReturnsErrorOnLoad(c *C) {
	s.repo.aggregateFactory = nil
	id := yooid()

	agg, err := s.repo.Load(typeOf(NewSomeAggregate(id)), id)

	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "The common domain repository has no Aggregate Factory.")
	c.Assert(agg, IsNil)
}

func (s *ComDomRepoSuite) TestRepositoryReturnsAnErrorIfAggregateFactoryNotRegisteredForAnAggregate(c *C) {
	aggregateFactory := NewDelegateAggregateFactory()
	aggregateFactory.RegisterDelegate(&SomeOtherAggregate{}, func(id uuid.UUID) AggregateRoot { return NewSomeOtherAggregate(id) })
	s.repo.SetAggregateFactory(aggregateFactory)

	id := yooid()
	aggregateTypeName := typeOf(&SomeAggregate{})
	agg, err := s.repo.Load(aggregateTypeName, id)

	c.Assert(err, DeepEquals,
		fmt.Errorf("The repository has no aggregate factory registered for aggregate type: %s",
			aggregateTypeName))
	c.Assert(agg, IsNil)
}

func (s *ComDomRepoSuite) TestCanRegisterStreamNameDelegate(c *C) {
	d := NewDelegateStreamNamer()

	s.repo.SetStreamNameDelegate(d)

	c.Assert(s.repo.streamNameDelegate, Equals, d)
}

func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfStreamNameDelegateNotRegisteredForAggregate(c *C) {
	id := yooid()
	streamNameDelegate := NewDelegateStreamNamer()
	streamNameDelegate.RegisterDelegate(func(t string, id uuid.UUID) string { return "something" },
		&SomeOtherAggregate{})
	s.repo.SetStreamNameDelegate(streamNameDelegate)
	typeName := typeOf(&SomeAggregate{})

	agg, err := s.repo.Load(typeName, id)

	c.Assert(agg, IsNil)
	c.Assert(err, DeepEquals,
		fmt.Errorf("There is no stream name delegate for aggregate of type \"%s\"",
			typeName))
}

func (s *ComDomRepoSuite) TestStreamNameIsBuiltByStreamNameDelegateOnSave(c *C) {
	id := yooid()
	agg := NewSomeAggregate(id)
	f := func(t string, id uuid.UUID) string { return "BoundedContext-" + id.String() }
	d := NewDelegateStreamNamer()
	d.RegisterDelegate(f, agg)
	s.repo.streamNameDelegate = d
	ev := NewTestEventMessage(id)
	agg.TrackChange(ev)

	err := s.repo.Save(agg)

	c.Assert(err, IsNil)
	c.Assert(s.store.stream, Equals, f(typeOf(agg), agg.AggregateID()))
}

func (s *ComDomRepoSuite) TestStreamNameIsBuiltByDelegateOnLoad(c *C) {
	id := yooid()
	agg := NewSomeAggregate(id)
	f := func(t string, id uuid.UUID) string { return "xyz-" + id.String() }
	d := NewDelegateStreamNamer()
	d.RegisterDelegate(f, agg)
	s.repo.streamNameDelegate = d

	_, err := s.repo.Load(typeOf(agg), agg.AggregateID())

	c.Assert(err, IsNil)
	c.Assert(s.store.stream, Equals, f(typeOf(agg), agg.AggregateID()))
}

func (s *ComDomRepoSuite) TestReturnsErrorOnSaveIfStreamNameDelegateNotRegisteredForAnAggregate(c *C) {
	streamNameDelegate := NewDelegateStreamNamer()
	streamNameDelegate.RegisterDelegate(func(t string, id uuid.UUID) string { return "something" })
	s.repo.SetStreamNameDelegate(streamNameDelegate)
	agg := NewSomeAggregate(yooid())

	err := s.repo.Save(agg)

	c.Assert(err, DeepEquals,
		fmt.Errorf("There is no stream name delegate for aggregate of type \"%s\"",
			typeOf(agg)))
}

func (s *ComDomRepoSuite) TestReturnsErrorOnSaveIfStreamNameDelegateIsNil(c *C) {
	s.repo.streamNameDelegate = nil
	agg := NewSomeAggregate(yooid())

	err := s.repo.Save(agg)

	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, fmt.Errorf("The common domain repository has no stream name delagate."))
}

func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfStreamNameDelegateIsNil(c *C) {
	s.repo.streamNameDelegate = nil

	_, err := s.repo.Load("", yooid())

	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, fmt.Errorf("The common domain repository has no stream name delegate."))
}

func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfEventFactoryNotRegistered(c *C) {
	s.repo.eventFactory = nil

	agg, err := s.repo.Load(typeOf(&SomeAggregate{}), yooid())

	c.Assert(err, DeepEquals, fmt.Errorf("The common domain has no Event Factory."))
	c.Assert(agg, IsNil)
}

func (s *ComDomRepoSuite) TestCanSetEventFactory(c *C) {
	eventFactory := NewDelegateEventFactory()

	s.repo.SetEventFactory(eventFactory)

	c.Assert(s.repo.eventFactory, Equals, eventFactory)
}

func (s *ComDomRepoSuite) TestAsyncFakeReadForwardAll(c *C) {

	stream := "some-stream"
	es := goes.CreateTestEvents(10, stream, "some-server", "MyEventType")
	ers := goes.CreateTestEventResponses(es, nil)
	fake := NewFakeAsyncClient()
	fake.eventResponses = ers

	eventsChannel := fake.ReadStreamForwardAsync(stream, nil, nil, 0)
	count := 0
	for {
		select {
		case ev, open := <-eventsChannel:
			if !open {
				c.Assert(count, Equals, len(es))
				c.Assert(fake.stream, Equals, stream)
				return
			}

			c.Assert(ev.Err, IsNil)
			c.Assert(ev.EventResp.Event, Equals, es[count])
			count++
		}
	}
}

func (s *ComDomRepoSuite) TestAppendToStream(c *C) {
	stream := "some-stream"
	es := goes.CreateTestEvents(10, stream, "some-server", "MyEventType")
	fake := NewFakeAsyncClient()

	resp, err := fake.AppendToStream(stream, nil, es...)

	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, http.StatusCreated)
	c.Assert(resp.StatusMessage, Equals, "201 Created")
	c.Assert(fake.appended, DeepEquals, es)
	c.Assert(fake.stream, Equals, stream)
}

func (s *ComDomRepoSuite) TestAggregateNotFoundError(c *C) {
	errorClient := &ErrorClient{}
	errorClient.resp = &goes.Response{StatusCode: http.StatusNotFound, StatusMessage: "404 Not Found"}
	errorClient.err = &goes.ErrorResponse{}
	s.repo.eventStore = errorClient

	id := yooid()
	agg, err := s.repo.Load(typeOf(&SomeAggregate{}), id)
	c.Assert(agg, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &AggregateNotFoundError{AggregateID: id, AggregateType: typeOf(&SomeAggregate{})})
}

func (s *ComDomRepoSuite) TestForwardEventStoreError(c *C) {
	errorClient := &ErrorClient{}
	errorClient.resp = &goes.Response{StatusCode: http.StatusUnauthorized, StatusMessage: "401 Unauthorized"}
	errorClient.err = fmt.Errorf("Some Error")
	s.repo.eventStore = errorClient

	id := yooid()
	agg, err := s.repo.Load(typeOf(&SomeAggregate{}), id)

	c.Assert(agg, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err, Equals, errorClient.err)
}

func (s *ComDomRepoSuite) TestSaveReturnsConncurrencyException(c *C) {
	errorClient := &ErrorClient{}
	errorClient.resp = &goes.Response{StatusCode: http.StatusBadRequest, StatusMessage: "400 Wrong expected EventNumber"}
	errorClient.err = &goes.ErrorResponse{}
	s.repo.eventStore = errorClient

	id := yooid()
	agg := NewSomeAggregate(id)
	agg.TrackChange(NewEventMessage(yooid(), &SomeEvent{"Some data", 4}))

	err := s.repo.Save(agg)

	c.Assert(err, FitsTypeOf, &ConcurrencyError{Aggregate: agg, ExpectedVersion: 1})
}

func (s *ComDomRepoSuite) TestSaveForwardsUnhandledErrors(c *C) {
	errorClient := &ErrorClient{}
	errorClient.resp = &goes.Response{StatusCode: http.StatusUnauthorized, StatusMessage: "401 Unauthorized"}
	errorClient.err = fmt.Errorf("Some Error")
	s.repo.eventStore = errorClient

	id := yooid()
	agg := NewSomeAggregate(id)
	agg.TrackChange(NewEventMessage(yooid(), &SomeEvent{"Some data", 4}))

	err := s.repo.Save(agg)

	c.Assert(err, Equals, errorClient.err)
}

//////////////////////////////////////////////////////////////////////////////
// Fakes

func NewStubAggregate(id uuid.UUID) *StubAggregate {

	return &StubAggregate{
		AggregateBase: NewAggregateBase(id),
	}
}

type StubAggregate struct {
	*AggregateBase
	events []EventMessage
}

func (t *StubAggregate) AggregateType() string {
	return "StubAggregate"
}

func (t *StubAggregate) Handle(command CommandMessage) error {
	return nil
}

func (t *StubAggregate) Apply(event EventMessage, isNew bool) {
	t.events = append(t.events, event)
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

func (c *FakeAsyncReader) ReadStreamForwardAsync(stream string, version *goes.StreamVersion, take *goes.Take, bufSize int) <-chan *goes.AsyncResponse {

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

type ErrorClient struct {
	err  error
	resp *goes.Response
}

func (c *ErrorClient) ReadStreamForwardAsync(stream string, version *goes.StreamVersion, take *goes.Take, bufSize int) <-chan *goes.AsyncResponse {

	eventsChannel := make(chan *goes.AsyncResponse)

	go func() {
		eventsChannel <- &goes.AsyncResponse{nil, c.resp, c.err}
		close(eventsChannel)
		return
	}()

	return eventsChannel

}

func (c *ErrorClient) AppendToStream(stream string, expectedVersion *goes.StreamVersion, events ...*goes.Event) (*goes.Response, error) {
	return c.resp, c.err
}
