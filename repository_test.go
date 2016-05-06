package ycq

import (
	//"fmt"
	"fmt"

	"github.com/jetbasrawi/yoono-uuid"
	. "gopkg.in/check.v1"
)

var _ = Suite(&ComDomRepoSuite{})

type ComDomRepoSuite struct {
	store *StubEventStore
	repo  *CommonDomainRepository
}

func (s *ComDomRepoSuite) SetUpTest(c *C) {
	s.store = &StubEventStore{
		events: make([]EventMessage, 0),
	}
	s.repo, _ = NewCommonDomainRepository(s.store)

	aggregateFactory := NewDelegateAggregateFactory()
	aggregateFactory.RegisterDelegate(&SomeAggregate{}, func(id uuid.UUID) AggregateRoot { return NewSomeAggregate(id) })
	s.repo.aggregateFactory = aggregateFactory

	streamNameDelegate := NewDelegateStreamNamer()
	streamNameDelegate.RegisterDelegate(func(at string, id uuid.UUID) string { return at + id.String() }, &SomeAggregate{})
	s.repo.streamNameDelegate = streamNameDelegate
}

func (s *ComDomRepoSuite) TestCanConstructNewRepository(c *C) {
	store := &StubEventStore{
		events: make([]EventMessage, 0),
	}
	repo, err := NewCommonDomainRepository(store)
	c.Assert(repo, NotNil)
	c.Assert(err, IsNil)
	c.Assert(repo.aggregateFactory, IsNil)
	c.Assert(repo.streamNameDelegate, IsNil)
}

func (s *ComDomRepoSuite) TestCreatingNewRepositoryWithNilEventStoreReturnsAnError(c *C) {
	repo, err := NewCommonDomainRepository(nil)
	c.Assert(repo, IsNil)
	c.Assert(err, DeepEquals, fmt.Errorf("Nil Eventstore injected into repository."))
}

func (s *ComDomRepoSuite) TestCanRegisterAggregateFactory(c *C) {
	aggregateFactory := NewDelegateAggregateFactory()

	err := s.repo.RegisterAggregateFactory(aggregateFactory)

	c.Assert(err, IsNil)
	c.Assert(s.repo.aggregateFactory, Equals, aggregateFactory)
}

func (s *ComDomRepoSuite) TestNoAggregateFactoryReturnsErrorOnLoad(c *C) {
	s.repo.aggregateFactory = nil
	id := yooid()

	agg, err := s.repo.Load(NewSomeAggregate(id).AggregateType(), id)

	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "The common domain repository has no Aggregate Factory.")
	c.Assert(agg, IsNil)
}

func (s *ComDomRepoSuite) TestRepositoryCanLoadAnAggregate(c *C) {
	id := yooid()
	agg, err := s.repo.Load(NewSomeAggregate(id).AggregateType(), id)

	c.Assert(err, IsNil)
	c.Assert(agg.AggregateID(), Equals, id)
	c.Assert(agg.AggregateType(), Equals, NewSomeAggregate(id).AggregateType())
	c.Assert(agg.Version(), Equals, 0)
}

func (s *ComDomRepoSuite) TestRepositoryCanLoadAggregateWithEvents(c *C) {
	af := NewDelegateAggregateFactory()
	_ = af.RegisterDelegate(&StubAggregate{},
		func(id uuid.UUID) AggregateRoot { return NewStubAggregate(id) })
	s.repo.RegisterAggregateFactory(af)
	id := yooid()
	agg := NewStubAggregate(id)
	ev := NewTestEventMessage(id)
	streamNameDelegate := NewDelegateStreamNamer()
	_ = streamNameDelegate.RegisterDelegate(func(t string, id uuid.UUID) string { return t + id.String() }, &StubAggregate{})
	s.repo.streamNameDelegate = streamNameDelegate
	stream, _ := s.repo.streamNameDelegate.GetStreamName(agg.AggregateType(), id)
	s.store.Save(stream, []EventMessage{ev}, nil, nil)

	got, err := s.repo.Load(NewStubAggregate(yooid()).AggregateType(), id)

	c.Assert(err, IsNil)
	c.Assert(got.AggregateID(), Equals, id)
	c.Assert(got.Version(), Equals, 1)
	c.Assert(got.(*StubAggregate).event, DeepEquals, ev)
}

func (s *ComDomRepoSuite) TestRepositoryIncrementsAggregateVersionForEachEvent(c *C) {
	af := NewDelegateAggregateFactory()
	_ = af.RegisterDelegate(&StubAggregate{},
		func(id uuid.UUID) AggregateRoot { return NewStubAggregate(id) })
	s.repo.RegisterAggregateFactory(af)
	id := yooid()
	agg := NewStubAggregate(id)
	streamNameDelegate := NewDelegateStreamNamer()
	_ = streamNameDelegate.RegisterDelegate(func(t string, id uuid.UUID) string { return t + id.String() }, &StubAggregate{})
	s.repo.streamNameDelegate = streamNameDelegate
	stream, _ := s.repo.streamNameDelegate.GetStreamName(agg.AggregateType(), id)
	ev1 := NewTestEventMessage(id)
	ev2 := NewTestEventMessage(id)
	ev3 := NewTestEventMessage(id)
	s.store.Save(stream, []EventMessage{ev1, ev2, ev3}, nil, nil)

	got, _ := s.repo.Load(NewStubAggregate(yooid()).AggregateType(), id)

	c.Assert(got.Version(), Equals, 3)
}

func (s *ComDomRepoSuite) TestSaveAggregateWithUncommittedChanges(c *C) {
	id := yooid()
	agg := NewSomeAggregate(id)
	ev := NewTestEventMessage(id)
	agg.StoreEvent(ev)

	err := s.repo.Save(agg)

	c.Assert(err, IsNil)
	stream, _ := s.repo.streamNameDelegate.GetStreamName(agg.AggregateType(), agg.AggregateID())
	events, err := s.store.Load(stream)
	c.Assert(err, IsNil)
	c.Assert(events, DeepEquals, []EventMessage{ev})
	c.Assert(agg.GetChanges(), DeepEquals, []EventMessage{})
	c.Assert(agg.Version(), Equals, 0)
}

func (s *ComDomRepoSuite) TestRepositoryReturnsAnErrorIfAggregateFactoryNotRegisteredForAnAggregate(c *C) {
	id := yooid()
	aggregateTypeName := NewSomeAggregate(yooid()).AggregateType()
	s.repo.RegisterAggregateFactory(NewDelegateAggregateFactory())

	agg, err := s.repo.Load(aggregateTypeName, id)

	c.Assert(err, DeepEquals,
		fmt.Errorf("The repository has no aggregate factory registered for aggregate type: %s",
			aggregateTypeName))
	c.Assert(agg, IsNil)
}

func (s *ComDomRepoSuite) TestStreamNameIsBuiltByStreamNameDelegateOnSave(c *C) {
	id := yooid()
	agg := NewSomeAggregate(id)
	f := func(t string, id uuid.UUID) string { return "BoundedContext-" + id.String() }
	d := NewDelegateStreamNamer()
	d.RegisterDelegate(f, agg)
	s.repo.streamNameDelegate = d
	ev := NewTestEventMessage(id)
	agg.StoreEvent(ev)

	err := s.repo.Save(agg)

	c.Assert(err, IsNil)
	c.Assert(s.store.stream, Equals, f(agg.AggregateType(), agg.AggregateID()))
}

func (s *ComDomRepoSuite) TestReturnsErrorOnSaveIfStreamNameDelegateNotRegistered(c *C) {
	agg := NewStubAggregate(yooid())

	err := s.repo.Save(agg)

	c.Assert(err, DeepEquals,
		fmt.Errorf("There is no stream name delegate for aggregate of type \"%s\"",
			agg.AggregateType()))
}

func (s *ComDomRepoSuite) TestReturnsErrorOnSaveIfStreamNameDelegateIsNil(c *C) {
	s.repo.streamNameDelegate = nil

	agg := NewSomeAggregate(yooid())
	err := s.repo.Save(agg)

	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, fmt.Errorf("The common domain repository has no stream name delagate."))
}

func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfStreamNameDelegateNotRegisteredForAggregate(c *C) {

	id := yooid()
	agg := NewSomeAggregate(id)

	s.repo.streamNameDelegate = NewDelegateStreamNamer()

	ev := NewTestEventMessage(id)
	s.store.Save(agg.AggregateID().String(), []EventMessage{ev}, nil, nil)
	_, err := s.repo.Load(agg.AggregateType(), agg.AggregateID())
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals,
		fmt.Errorf("There is no stream name delegate for aggregate of type \"%s\"",
			agg.AggregateType()))
}

func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfStreamNameDelegateIsNil(c *C) {
	s.repo.streamNameDelegate = nil

	_, err := s.repo.Load("", yooid())

	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, fmt.Errorf("The common domain repository has no stream name delegate."))
}

func (s *ComDomRepoSuite) TestStreamNameIsBuiltByDelegateOnLoad(c *C) {
	id := yooid()
	agg := NewSomeAggregate(id)
	ev := NewTestEventMessage(id)
	f := func(t string, id uuid.UUID) string { return "xyz-" + id.String() }
	d := NewDelegateStreamNamer()
	d.RegisterDelegate(f, agg)
	s.repo.streamNameDelegate = d
	s.store.Save(f(agg.AggregateType(), agg.AggregateID()), []EventMessage{ev}, nil, nil)

	_, err := s.repo.Load(agg.AggregateType(), agg.AggregateID())

	c.Assert(err, IsNil)
	c.Assert(s.store.loaded, Equals, f(agg.AggregateType(), agg.AggregateID()))
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
	event EventMessage
}

func (t *StubAggregate) AggregateType() string {
	return "StubAggregate"
}

func (t *StubAggregate) Handle(command CommandMessage) error {
	return nil
}

func (t *StubAggregate) Apply(event EventMessage) {
	t.event = event
}
