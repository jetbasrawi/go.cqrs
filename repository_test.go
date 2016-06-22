package ycq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/jetbasrawi/goes"
	"github.com/satori/go.uuid"
	. "gopkg.in/check.v1"
)

var (
	_ = Suite(&ComDomRepoSuite{})
)

func (s *ComDomRepoSuite) SetupSimulator(es []*goes.Event, m *goes.Event) {
	s.stubFeed.Events = es
	s.stubFeed.MetaData = m
}

type ComDomRepoSuite struct {
	repo           *CommonDomainRepository
	someEvent      *SomeEvent
	someMeta       map[string]string
	goesEvent      *goes.Event
	someOtherEvent *SomeOtherEvent
	someOtherMeta  map[string]string
	goesOtherEvent *goes.Event
	mux            *http.ServeMux
	server         *httptest.Server
	client         *goes.Client
	streamName     string
	stubFeed       *goes.ESAtomFeedSimulator
}

func (s *ComDomRepoSuite) SetUpTest(c *C) {
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)
	s.client, _ = goes.NewClient(nil, s.server.URL)

	s.streamName = "astream"
	s.someEvent = &SomeEvent{Item: "Some Item", Count: 42}
	s.someMeta = map[string]string{"AggregateID": uuid.NewV4().String()}
	s.goesEvent = goes.CreateTestEventFromData(s.streamName, s.server.URL, 0, s.someEvent, s.someMeta)
	s.someOtherEvent = &SomeOtherEvent{yooid()}
	s.someOtherMeta = map[string]string{"AggregateID": uuid.NewV4().String()}
	s.goesOtherEvent = goes.CreateTestEventFromData(s.streamName, s.server.URL, 1, s.someOtherEvent, s.someOtherMeta)
	es := []*goes.Event{s.goesEvent, s.goesOtherEvent}

	u, _ := url.Parse(s.server.URL)
	s.stubFeed = &goes.ESAtomFeedSimulator{Events: es, BaseURL: u, MetaData: nil}
	s.mux.Handle("/", s.stubFeed)
	s.SetupSimulator(es, nil)

	eventBus := NewInternalEventBus()

	s.repo, _ = NewCommonDomainRepository(s.client, eventBus)

	aggregateFactory := NewDelegateAggregateFactory()
	aggregateFactory.RegisterDelegate(&SomeAggregate{},
		func(id string) AggregateRoot { return NewSomeAggregate(id) })
	s.repo.aggregateFactory = aggregateFactory

	streamNameDelegate := NewDelegateStreamNamer()
	streamNameDelegate.RegisterDelegate(func(at string, id string) string { return at + id },
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

//TODO - Fix up feed with no entries
//func (s *ComDomRepoSuite) TestRepositoryCanLoadAnAggregate(c *C) {
//s.SetupSimulator([]*goes.Event{}, nil)
//id := yooid()
//agg, err := s.repo.Load(typeOf(&SomeAggregate{}), id)

//c.Assert(err, IsNil)
//c.Assert(agg.AggregateID(), Equals, id)
//c.Assert(typeOf(agg), Equals, typeOf(NewSomeAggregate(id)))
//c.Assert(agg.Version(), Equals, 0)
//}

func (s *ComDomRepoSuite) TestRepositoryCanLoadAggregateWithEvents(c *C) {
	id := yooid()
	aggregateFactory := NewDelegateAggregateFactory()
	aggregateFactory.RegisterDelegate(&StubAggregate{},
		func(id string) AggregateRoot { return NewStubAggregate(id) })
	s.repo.SetAggregateFactory(aggregateFactory)

	got, err := s.repo.Load(typeOf(&StubAggregate{}), id)

	c.Assert(err, IsNil)
	c.Assert(got.AggregateID(), Equals, id)

	events := got.(*StubAggregate).events

	c.Assert(events[0].Event(), DeepEquals, s.someEvent)
	c.Assert(events[1].Event(), DeepEquals, s.someOtherEvent)
}

func (s *ComDomRepoSuite) TestRepositoryIncrementsAggregateVersionForEachEvent(c *C) {
	ev1 := goes.CreateTestEventFromData(s.streamName, s.server.URL, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	ev2 := goes.CreateTestEventFromData(s.streamName, s.server.URL, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	ev3 := goes.CreateTestEventFromData(s.streamName, s.server.URL, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	es := []*goes.Event{ev1, ev2, ev3}

	s.SetupSimulator(es, nil)

	id := yooid()
	got, _ := s.repo.Load(typeOf(&SomeAggregate{}), id)

	c.Assert(got.Version(), Equals, 3)
}

//func (s *ComDomRepoSuite) TestSaveAggregateWithUncommittedChanges(c *C) {
//s.stream.eventResponses = []*goes.EventResponse{}
//id := yooid()
//agg := NewSomeAggregate(id)
//ev := &SomeEvent{Item: "Some string", Count: 4353}
//em := NewEventMessage(id, ev)
//agg.TrackChange(em)

//err := s.repo.Save(agg)

//c.Assert(err, IsNil)
//c.Assert(s.stream.appended[0].Data, DeepEquals, ev)
//}

//func (s *ComDomRepoSuite) TestCanRegisterAggregateFactory(c *C) {
//aggregateFactory := NewDelegateAggregateFactory()

//s.repo.SetAggregateFactory(aggregateFactory)

//c.Assert(s.repo.aggregateFactory, Equals, aggregateFactory)
//}

//func (s *ComDomRepoSuite) TestNoAggregateFactoryReturnsErrorOnLoad(c *C) {
//s.repo.aggregateFactory = nil
//id := yooid()

//agg, err := s.repo.Load(typeOf(NewSomeAggregate(id)), id)

//c.Assert(err, NotNil)
//c.Assert(err, ErrorMatches, "The common domain repository has no Aggregate Factory.")
//c.Assert(agg, IsNil)
//}

//func (s *ComDomRepoSuite) TestRepositoryReturnsAnErrorIfAggregateFactoryNotRegisteredForAnAggregate(c *C) {
//aggregateFactory := NewDelegateAggregateFactory()
//aggregateFactory.RegisterDelegate(&SomeOtherAggregate{}, func(id string) AggregateRoot { return NewSomeOtherAggregate(id) })
//s.repo.SetAggregateFactory(aggregateFactory)

//id := yooid()
//aggregateTypeName := typeOf(&SomeAggregate{})
//agg, err := s.repo.Load(aggregateTypeName, id)

//c.Assert(err, DeepEquals,
//fmt.Errorf("The repository has no aggregate factory registered for aggregate type: %s",
//aggregateTypeName))
//c.Assert(agg, IsNil)
//}

//func (s *ComDomRepoSuite) TestCanRegisterStreamNameDelegate(c *C) {
//d := NewDelegateStreamNamer()

//s.repo.SetStreamNameDelegate(d)

//c.Assert(s.repo.streamNameDelegate, Equals, d)
//}

//func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfStreamNameDelegateNotRegisteredForAggregate(c *C) {
//id := yooid()
//streamNameDelegate := NewDelegateStreamNamer()
//streamNameDelegate.RegisterDelegate(func(t string, id string) string { return "something" },
//&SomeOtherAggregate{})
//s.repo.SetStreamNameDelegate(streamNameDelegate)
//typeName := typeOf(&SomeAggregate{})

//agg, err := s.repo.Load(typeName, id)

//c.Assert(agg, IsNil)
//c.Assert(err, DeepEquals,
//fmt.Errorf("There is no stream name delegate for aggregate of type \"%s\"",
//typeName))
//}

//func (s *ComDomRepoSuite) TestStreamNameIsBuiltByStreamNameDelegateOnSave(c *C) {
//id := yooid()
//agg := NewSomeAggregate(id)
//f := func(t string, id string) string { return "BoundedContext-" + id }
//d := NewDelegateStreamNamer()
//d.RegisterDelegate(f, agg)
//s.repo.streamNameDelegate = d
//ev := NewTestEventMessage(id)
//agg.TrackChange(ev)

//err := s.repo.Save(agg)

//c.Assert(err, IsNil)
//c.Assert(s.stream.streamName, Equals, f(typeOf(agg), agg.AggregateID()))
//}

//func (s *ComDomRepoSuite) TestStreamNameIsBuiltByDelegateOnLoad(c *C) {
//id := yooid()
//agg := NewSomeAggregate(id)
//f := func(t string, id string) string { return "xyz-" + id }
//d := NewDelegateStreamNamer()
//d.RegisterDelegate(f, agg)
//s.repo.streamNameDelegate = d

//_, err := s.repo.Load(typeOf(agg), agg.AggregateID())

//c.Assert(err, IsNil)
//c.Assert(s.stream.streamName, Equals, f(typeOf(agg), agg.AggregateID()))
//}

//func (s *ComDomRepoSuite) TestReturnsErrorOnSaveIfStreamNameDelegateNotRegisteredForAnAggregate(c *C) {
//streamNameDelegate := NewDelegateStreamNamer()
//streamNameDelegate.RegisterDelegate(func(t string, id string) string { return "something" })
//s.repo.SetStreamNameDelegate(streamNameDelegate)
//agg := NewSomeAggregate(yooid())

//err := s.repo.Save(agg)

//c.Assert(err, DeepEquals,
//fmt.Errorf("There is no stream name delegate for aggregate of type \"%s\"",
//typeOf(agg)))
//}

//func (s *ComDomRepoSuite) TestReturnsErrorOnSaveIfStreamNameDelegateIsNil(c *C) {
//s.repo.streamNameDelegate = nil
//agg := NewSomeAggregate(yooid())

//err := s.repo.Save(agg)

//c.Assert(err, NotNil)
//c.Assert(err, DeepEquals, fmt.Errorf("The common domain repository has no stream name delagate."))
//}

//func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfStreamNameDelegateIsNil(c *C) {
//s.repo.streamNameDelegate = nil

//_, err := s.repo.Load("", yooid())

//c.Assert(err, NotNil)
//c.Assert(err, DeepEquals, fmt.Errorf("The common domain repository has no stream name delegate."))
//}

//func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfEventFactoryNotRegistered(c *C) {
//s.repo.eventFactory = nil

//agg, err := s.repo.Load(typeOf(&SomeAggregate{}), yooid())

//c.Assert(err, DeepEquals, fmt.Errorf("The common domain has no Event Factory."))
//c.Assert(agg, IsNil)
//}

//func (s *ComDomRepoSuite) TestCanSetEventFactory(c *C) {
//eventFactory := NewDelegateEventFactory()

//s.repo.SetEventFactory(eventFactory)

//c.Assert(s.repo.eventFactory, Equals, eventFactory)
//}

//func (s *ComDomRepoSuite) TestAsyncFakeReadForwardAll(c *C) {

//stream := "some-stream"
//es := goes.CreateTestEvents(10, stream, "some-server", "MyEventType")
//ers := goes.CreateTestEventResponses(es, nil)
//fake := NewFakeAsyncClient()
//fake.eventResponses = ers

//client := NewFakeClient(fake)

//eventsChannel := fake.ReadStreamForwardAsync(stream, nil, nil, 0)
//count := 0
//for {
//select {
//case ev, open := <-eventsChannel:
//if !open {
//c.Assert(count, Equals, len(es))
//c.Assert(fake.stream, Equals, stream)
//return
//}

//c.Assert(ev.Err, IsNil)
//c.Assert(ev.EventResp.Event, Equals, es[count])
//count++
//}
//}
//}

//func (s *ComDomRepoSuite) TestAppendToStream(c *C) {
//stream := "some-stream"
//es := goes.CreateTestEvents(10, stream, "some-server", "MyEventType")
//fake := NewFakeAsyncClient()
//client := NewFakeClient(fake)

//resp, err := client.AppendToStream(stream, nil, es...)

//c.Assert(err, IsNil)
//c.Assert(resp.StatusCode, Equals, http.StatusCreated)
//c.Assert(resp.StatusMessage, Equals, "201 Created")
//c.Assert(fake.appended, DeepEquals, es)
//c.Assert(fake.stream, Equals, stream)
//}

//TODO FIX
//func (s *ComDomRepoSuite) TestAggregateNotFoundError(c *C) {
//errorClient := &ErrorClient{}
//errorClient.resp = &goes.Response{StatusCode: http.StatusNotFound, StatusMessage: "404 Not Found"}
//errorClient.err = &goes.ErrorResponse{}
//s.repo.eventStore = errorClient

//id := yooid()
//agg, err := s.repo.Load(typeOf(&SomeAggregate{}), id)
//c.Assert(agg, IsNil)
//c.Assert(err, NotNil)
//c.Assert(err, FitsTypeOf, &AggregateNotFoundError{AggregateID: id, AggregateType: typeOf(&SomeAggregate{})})
//}

//TODO FIX
//func (s *ComDomRepoSuite) TestForwardEventStoreError(c *C) {
//errorClient := &ErrorClient{}
//errorClient.resp = &goes.Response{StatusCode: http.StatusUnauthorized, StatusMessage: "401 Unauthorized"}
//errorClient.err = fmt.Errorf("Some Error")
//s.repo.eventStore = errorClient

//id := yooid()
//agg, err := s.repo.Load(typeOf(&SomeAggregate{}), id)

//c.Assert(agg, IsNil)
//c.Assert(err, NotNil)
//c.Assert(err, Equals, errorClient.err)
//}

//TODO Fix
//func (s *ComDomRepoSuite) TestSaveReturnsConncurrencyException(c *C) {
//errorClient := &ErrorClient{}
//errorClient.resp = &goes.Response{StatusCode: http.StatusBadRequest, StatusMessage: "400 Wrong expected EventNumber"}
//errorClient.err = &goes.ErrorResponse{}
//s.repo.eventStore = errorClient

//id := yooid()
//agg := NewSomeAggregate(id)
//agg.TrackChange(NewEventMessage(yooid(), &SomeEvent{"Some data", 4}))

//err := s.repo.Save(agg)

//c.Assert(err, FitsTypeOf, &ConcurrencyError{Aggregate: agg, ExpectedVersion: 1})
//}

//TODO Fix
//func (s *ComDomRepoSuite) TestSaveForwardsUnhandledErrors(c *C) {
//errorClient := &ErrorClient{}
//errorClient.resp = &goes.Response{StatusCode: http.StatusUnauthorized, StatusMessage: "401 Unauthorized"}
//errorClient.err = fmt.Errorf("Some Error")
//s.repo.eventStore = errorClient

//id := yooid()
//agg := NewSomeAggregate(id)
//agg.TrackChange(NewEventMessage(yooid(), &SomeEvent{"Some data", 4}))

//err := s.repo.Save(agg)

//c.Assert(err, Equals, errorClient.err)
//}

//////////////////////////////////////////////////////////////////////////////
// Fakes

func NewStubAggregate(id string) *StubAggregate {
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

//type FakeClient struct {
//stream *FakeStream
//}

//func NewFakeClient(stream *FakeStream) *FakeClient {
//return &FakeClient{
//stream: stream,
//}
//}

//func (c *FakeClient) Dial(s string) goes.StreamReader {
//return c.stream
//}

//func (c *FakeClient) AppendToStream(stream string, expectedVersion *goes.StreamVersion, events ...*goes.Event) (*goes.Response, error) {
//c.stream.appended = events
//c.stream.streamName = stream
//r := &goes.Response{
//StatusCode:    http.StatusCreated,
//StatusMessage: "201 Created",
//}
//return r, nil
//}

//type FakeStream struct {
//eventResponses []*goes.EventResponse
//streamName     string
//appended       []*goes.Event
//version        int
//}

//func NewFakeStream() *FakeStream {
//fake := &FakeStream{
//version: -1,
//}
//return fake
//}

//func (c *FakeStream) Version() int {
//return c.version
//}

//func (c *FakeStream) Err() error {
//return nil
//}

//func (c *FakeStream) Next() bool {
//c.version++
//return true
//}

//func (c *FakeStream) EventResponse() *goes.EventResponse {
//return nil
//}

//func (c *FakeStream) Scan(e interface{}) error {

//fmt.Printf("Vers: %d", c.Version())

//data, ok := c.eventResponses[c.version].Event.Data.(*json.RawMessage)
//if !ok {
//return fmt.Errorf("Could not unmarshal the event. Event data is not of type *json.RawMessage")
//}
//if err := json.Unmarshal(*data, e); err != nil {
//return err
//}
//return nil
//}

//type FakeAsyncReader struct {
//eventResponses []*goes.EventResponse
//stream         string
//appended       []*goes.Event
//}

//func NewFakeAsyncClient() *FakeAsyncReader {
//fake := &FakeAsyncReader{}
//return fake
//}

//func (c *FakeAsyncReader) ReadStreamForwardAsync(stream string, version *goes.StreamVersion, take *goes.Take, bufSize int) <-chan *goes.AsyncResponse {

//c.stream = stream
//eventsChannel := make(chan *goes.AsyncResponse)

//go func() {
//for _, v := range c.eventResponses {
//eventsChannel <- &goes.AsyncResponse{v, nil, nil}
//}
//close(eventsChannel)
//return
//}()

//return eventsChannel

//}

//func (c *FakeAsyncReader) AppendToStream(stream string, expectedVersion *goes.StreamVersion, events ...*goes.Event) (*goes.Response, error) {
//c.appended = events
//c.stream = stream
//r := &goes.Response{
//StatusCode:    http.StatusCreated,
//StatusMessage: "201 Created",
//}
//return r, nil
//}

//type ErrorClient struct {
//err  error
//resp *goes.Response
//}

//func (c *ErrorClient) ReadStreamForwardAsync(stream string, version *goes.StreamVersion, take *goes.Take, bufSize int) <-chan *goes.AsyncResponse {

//eventsChannel := make(chan *goes.AsyncResponse)

//go func() {
//eventsChannel <- &goes.AsyncResponse{nil, c.resp, c.err}
//close(eventsChannel)
//return
//}()

//return eventsChannel

//}

//func (c *ErrorClient) AppendToStream(stream string, expectedVersion *goes.StreamVersion, events ...*goes.Event) (*goes.Response, error) {
//return c.resp, c.err
//}
