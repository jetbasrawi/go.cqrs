// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/jetbasrawi/go.geteventstore"
	"github.com/jetbasrawi/go.geteventstore.testfeed"
	. "gopkg.in/check.v1"
)

var (
	_ = Suite(&ComDomRepoSuite{})
)

type ComDomRepoSuite struct {
	eventBus           EventBus
	repo               *GetEventStoreCommonDomainRepo
	someEvent          *SomeEvent
	someMeta           map[string]string
	someMockEvent      *mock.Event
	someOtherEvent     *SomeOtherEvent
	someOtherMeta      map[string]string
	someOtherMockEvent *mock.Event
	mux                *http.ServeMux
	server             *httptest.Server
	client             *goes.Client
	streamName         string
	stubFeed           *mock.AtomFeedSimulator
}

func (s *ComDomRepoSuite) SetUpTest(c *C) {
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)
	s.client, _ = goes.NewClient(nil, s.server.URL)

	s.SetupDefaultRepo(s.client)
}

func (s *ComDomRepoSuite) TearDownTest(c *C) {
	s.server.Close()
}

func (s *ComDomRepoSuite) SetupDefaultSimulator() {
	s.streamName = "astream"

	// Set up an event of type SomeEvent
	s.someEvent = &SomeEvent{Item: "Some Item", Count: 42}
	s.someMeta = map[string]string{"AggregateID": NewUUID()}
	s.someMockEvent = mock.CreateTestEventFromData(s.streamName, s.server.URL, 0, s.someEvent, s.someMeta)

	// Set up an event of type SomeOtherEvent
	s.someOtherEvent = &SomeOtherEvent{NewUUID()}
	s.someOtherMeta = map[string]string{"AggregateID": NewUUID()}
	s.someOtherMockEvent = mock.CreateTestEventFromData(s.streamName, s.server.URL, 1, s.someOtherEvent, s.someOtherMeta)

	// Create a slice with the two events for the Atom feed simulator
	es := []*mock.Event{s.someMockEvent, s.someOtherMockEvent}

	s.SetupSimulator(es, nil)
}

func (s *ComDomRepoSuite) SetupSimulator(es []*mock.Event, m *mock.Event) {
	// Set up an AtomFeedSimulator
	u, _ := url.Parse(s.server.URL)
	sim, err := mock.NewAtomFeedSimulator(es, u, nil, -1)
	if err != nil {
		log.Fatal(err)
	}
	s.stubFeed = sim

	// Set the http handler
	s.mux.Handle("/", s.stubFeed)
}

func (s *ComDomRepoSuite) SetupDefaultRepo(client *goes.Client) {
	s.eventBus = NewInternalEventBus()

	s.repo, _ = NewCommonDomainRepository(client, s.eventBus)

	aggregateFactory := NewDelegateAggregateFactory()
	aggregateFactory.RegisterDelegate(&SomeAggregate{},
		func(id string) AggregateRoot { return NewSomeAggregate(id) })
	s.repo.aggregateFactory = aggregateFactory

	streamNameDelegate := NewDelegateStreamNamer()
	streamNameDelegate.RegisterDelegate(func(at string, id string) string { return at + "-" + id },
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

func (s *ComDomRepoSuite) TestRepositoryCanLoadAggregateWithEvents(c *C) {

	s.SetupDefaultSimulator()

	id := NewUUID()
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
	ev1 := mock.CreateTestEventFromData(s.streamName, s.server.URL, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	ev2 := mock.CreateTestEventFromData(s.streamName, s.server.URL, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	ev3 := mock.CreateTestEventFromData(s.streamName, s.server.URL, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	es := []*mock.Event{ev1, ev2, ev3}

	s.SetupSimulator(es, nil)

	id := NewUUID()
	got, err := s.repo.Load(typeOf(&SomeAggregate{}), id)
	c.Assert(err, IsNil)

	// Version is a zero based index. The first item is zero
	c.Assert(got.OriginalVersion(), Equals, 2)
	c.Assert(got.CurrentVersion(), Equals, 2)
}

func (s *ComDomRepoSuite) TestSaveAggregateWithUncommittedChanges(c *C) {

	someEvent := &SomeEvent{Item: "Some string", Count: 4353}
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)

		es := []*goes.Event{}
		var d json.RawMessage
		var m json.RawMessage
		ev := &goes.Event{Data: &d, MetaData: &m}
		es = append(es, ev)

		err := json.NewDecoder(r.Body).Decode(&es)
		c.Assert(err, IsNil)

		data, ok := es[0].Data.(*json.RawMessage)
		c.Assert(ok, Equals, true)
		e := &SomeEvent{}
		err = json.Unmarshal(*data, e)
		c.Assert(e, DeepEquals, someEvent)

	})

	id := NewUUID()
	agg := NewSomeAggregate(id)

	em := NewEventMessage(id, someEvent, nil)
	agg.TrackChange(em)

	err := s.repo.Save(agg, nil)

	c.Assert(err, IsNil)

}

func (s *ComDomRepoSuite) TestCanRegisterAggregateFactory(c *C) {
	aggregateFactory := NewDelegateAggregateFactory()

	s.repo.SetAggregateFactory(aggregateFactory)

	c.Assert(s.repo.aggregateFactory, Equals, aggregateFactory)
}

func (s *ComDomRepoSuite) TestNoAggregateFactoryReturnsErrorOnLoad(c *C) {
	s.repo.aggregateFactory = nil
	id := NewUUID()

	agg, err := s.repo.Load(typeOf(NewSomeAggregate(id)), id)

	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "The common domain repository has no Aggregate Factory.")
	c.Assert(agg, IsNil)
}

func (s *ComDomRepoSuite) TestRepositoryReturnsAnErrorIfAggregateFactoryNotRegisteredForAnAggregate(c *C) {
	aggregateFactory := NewDelegateAggregateFactory()
	aggregateFactory.RegisterDelegate(&SomeOtherAggregate{}, func(id string) AggregateRoot { return NewSomeOtherAggregate(id) })
	s.repo.SetAggregateFactory(aggregateFactory)

	id := NewUUID()
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
	id := NewUUID()
	streamNameDelegate := NewDelegateStreamNamer()
	streamNameDelegate.RegisterDelegate(func(t string, id string) string { return "something" },
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

	id := NewUUID()
	agg := NewSomeAggregate(id)
	f := func(t string, id string) string { return "BoundedContext-" + id }
	d := NewDelegateStreamNamer()
	d.RegisterDelegate(f, agg)
	s.repo.streamNameDelegate = d
	ev := NewTestEventMessage(id)
	agg.TrackChange(ev)

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)
		streamName := f("", id)

		c.Assert(r.URL.String(), DeepEquals, fmt.Sprintf("/streams/%s", streamName))
	})

	err := s.repo.Save(agg, nil)

	c.Assert(err, IsNil)
}

func (s *ComDomRepoSuite) TestStreamNameIsBuiltByDelegateOnLoad(c *C) {
	id := NewUUID()
	agg := NewSomeAggregate(id)
	f := func(t string, id string) string { return "xyz-" + id }
	d := NewDelegateStreamNamer()
	d.RegisterDelegate(f, agg)
	s.repo.streamNameDelegate = d

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodGet)
		streamName := f("", id)
		c.Assert(r.URL.String(), DeepEquals, fmt.Sprintf("/streams/%s/0/forward/20", streamName))
	})

	_, _ = s.repo.Load(typeOf(agg), agg.AggregateID())

}

func (s *ComDomRepoSuite) TestReturnsErrorOnSaveIfStreamNameDelegateNotRegisteredForAnAggregate(c *C) {
	streamNameDelegate := NewDelegateStreamNamer()
	streamNameDelegate.RegisterDelegate(func(t string, id string) string { return "something" })
	s.repo.SetStreamNameDelegate(streamNameDelegate)
	agg := NewSomeAggregate(NewUUID())

	err := s.repo.Save(agg, nil)

	c.Assert(err, DeepEquals,
		fmt.Errorf("There is no stream name delegate for aggregate of type \"%s\"",
			typeOf(agg)))
}

func (s *ComDomRepoSuite) TestReturnsErrorOnSaveIfStreamNameDelegateIsNil(c *C) {
	s.repo.streamNameDelegate = nil
	agg := NewSomeAggregate(NewUUID())

	err := s.repo.Save(agg, nil)

	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, fmt.Errorf("The common domain repository has no stream name delagate."))
}

func (s *ComDomRepoSuite) TestLoadReturnErrUnauthorized(c *C) {
	id := NewUUID()
	agg := NewSomeAggregate(id)

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodGet)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "")
	})

	_, err := s.repo.Load(typeOf(agg), agg.AggregateID())
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrUnauthorized{})

}

func (s *ComDomRepoSuite) TestLoadReturnErrUnavailable(c *C) {
	id := NewUUID()
	agg := NewSomeAggregate(id)

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodGet)
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "")
	})

	_, err := s.repo.Load(typeOf(agg), agg.AggregateID())
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrRepositoryUnavailable{})

}

func (s *ComDomRepoSuite) TestSaveReturnErrUnauthorized(c *C) {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "")
	})

	id := NewUUID()
	agg := NewSomeAggregate(id)
	agg.TrackChange(NewEventMessage(NewUUID(), &SomeEvent{"Some data", 4}, nil))

	err := s.repo.Save(agg, nil)
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrUnauthorized{})

}

func (s *ComDomRepoSuite) TestSaveReturnErrUnavailable(c *C) {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "")
	})

	id := NewUUID()
	agg := NewSomeAggregate(id)
	agg.TrackChange(NewEventMessage(NewUUID(), &SomeEvent{"Some data", 4}, nil))

	err := s.repo.Save(agg, nil)
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrRepositoryUnavailable{})

}

func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfStreamNameDelegateIsNil(c *C) {
	s.repo.streamNameDelegate = nil

	_, err := s.repo.Load("", NewUUID())

	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, fmt.Errorf("The common domain repository has no stream name delegate."))
}

func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfEventFactoryNotRegistered(c *C) {
	s.repo.eventFactory = nil

	agg, err := s.repo.Load(typeOf(&SomeAggregate{}), NewUUID())

	c.Assert(err, DeepEquals, fmt.Errorf("The common domain has no Event Factory."))
	c.Assert(agg, IsNil)
}

func (s *ComDomRepoSuite) TestCanSetEventFactory(c *C) {
	eventFactory := NewDelegateEventFactory()

	s.repo.SetEventFactory(eventFactory)

	c.Assert(s.repo.eventFactory, Equals, eventFactory)
}

func (s *ComDomRepoSuite) TestAggregateNotFoundError(c *C) {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodGet)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "")
	})

	id := NewUUID()
	agg, err := s.repo.Load(typeOf(&SomeAggregate{}), id)
	c.Assert(agg, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrAggregateNotFound{AggregateID: id, AggregateType: typeOf(&SomeAggregate{})})
}

func (s *ComDomRepoSuite) TestSaveReturnsConncurrencyException(c *C) {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "")
	})

	id := NewUUID()
	agg := NewSomeAggregate(id)
	agg.TrackChange(NewEventMessage(NewUUID(), &SomeEvent{"Some data", 4}, nil))

	err := s.repo.Save(agg, Int(-1))

	streamName, _ := s.repo.streamNameDelegate.GetStreamName(typeOf(agg), id)
	c.Assert(err, DeepEquals, &ErrConcurrencyViolation{Aggregate: agg, ExpectedVersion: Int(-1), StreamName: streamName})
}

func (s *ComDomRepoSuite) TestSaveUnhandledErrors(c *C) {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, "")
	})

	id := NewUUID()
	agg := NewSomeAggregate(id)
	agg.TrackChange(NewEventMessage(NewUUID(), &SomeEvent{"Some data", 4}, nil))

	err := s.repo.Save(agg, nil)
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrUnexpected{})

}

func (s *ComDomRepoSuite) TestNewEventsArePublishedOnSave(c *C) {
	fakeHandler := &FakeEventHandler{}
	s.eventBus.AddHandler(fakeHandler, &SomeEvent{}, &SomeOtherEvent{})

	em1 := NewEventMessage(NewUUID(), &SomeEvent{"--------PUBLISH", 456}, nil)
	em2 := NewEventMessage(NewUUID(), &SomeOtherEvent{"--------PUBLISH"}, nil)

	agg := NewSomeAggregate(NewUUID())
	agg.TrackChange(em1)
	agg.TrackChange(em2)

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, "")
	})

	err := s.repo.Save(agg, Int(agg.OriginalVersion()))
	c.Assert(err, IsNil)

	//spew.Dump(s.eventBus)

	c.Assert(fakeHandler.Events, HasLen, 2)
	got1 := fakeHandler.Events[0].Version()
	got2 := fakeHandler.Events[1].Version()
	c.Assert(*got1, Equals, 0)
	c.Assert(*got2, Equals, 1)

}

//////////////////////////////////////////////////////////////////////////////
// Fakes

type FakeEventHandler struct {
	Events []EventMessage
}

func (h *FakeEventHandler) Handle(message EventMessage) {
	h.Events = append(h.Events, message)
}

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
