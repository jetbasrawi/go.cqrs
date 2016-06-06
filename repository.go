package ycq

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jetbasrawi/goes"
)

//GetEventStoreRepositoryClient is an interface that reflects the goes client.
type GetEventStoreRepositoryClient interface {
	ReadStreamForwardAsync(string, *goes.StreamVersion, *goes.Take, int) <-chan *goes.AsyncResponse
	AppendToStream(string, *goes.StreamVersion, ...*goes.Event) (*goes.Response, error)
}

//AggregateNotFoundError error returned when an aggregate was not found in the repository.
type AggregateNotFoundError struct {
	AggregateID   string
	AggregateType string
}

func (e *AggregateNotFoundError) Error() string {
	return fmt.Sprintf("Could not find any aggregate of type %s with id %s",
		e.AggregateType,
		e.AggregateID)
}

//ConcurrencyError is returned when a concurrency error is raised by the event store
//when events are persisted to a stream and the version of the stream does not match
//the expected version.
type ConcurrencyError struct {
	Aggregate       AggregateRoot
	ExpectedVersion int
	StreamName      string
}

func (e *ConcurrencyError) Error() string {
	return fmt.Sprintf("ConcurrencyError: AggregateID: %s ExpectedVersion: %d StreamName: %s", e.Aggregate.AggregateID(), e.ExpectedVersion, e.StreamName)
}

//DomainRepository is the interface that all domain repositories should implement.
type DomainRepository interface {
	//Loads an aggregate of the given type and ID
	Load(string, string) (AggregateRoot, error)

	//Saves the aggregate.
	Save(AggregateRoot) error
}

//CommonDomainRepository is a generic repository implementation
type CommonDomainRepository struct {
	eventStore         GetEventStoreRepositoryClient
	eventBus           EventBus
	streamNameDelegate StreamNamer
	aggregateFactory   AggregateFactory
	eventFactory       EventFactory
}

//NewCommonDomainRepository constructs a new CommonDomainRepository
func NewCommonDomainRepository(eventStore GetEventStoreRepositoryClient, eventBus EventBus) (*CommonDomainRepository, error) {
	if eventStore == nil {
		return nil, fmt.Errorf("Nil Eventstore injected into repository.")
	}

	if eventBus == nil {
		return nil, fmt.Errorf("Nil EventBus injected into repository.")
	}

	d := &CommonDomainRepository{
		eventStore: eventStore,
		eventBus:   eventBus,
	}
	return d, nil
}

//SetAggregateFactory sets the aggregate factory that should be used to
//instantate aggregate instances
//
//Only one AggregateFactory can be registered at any one time.
//Any registration will overwrite the provious registration.
func (r *CommonDomainRepository) SetAggregateFactory(factory AggregateFactory) {
	r.aggregateFactory = factory
}

//SetEventFactory sets the event factory that should be used to instantiate event
//instances.
//
//Only one event factory can be set at a time. Any subsequent registration will
//overwrite the previous factory.
func (r *CommonDomainRepository) SetEventFactory(factory EventFactory) {
	r.eventFactory = factory
}

//SetStreamNameDelegate sets the stream name delegate
func (r *CommonDomainRepository) SetStreamNameDelegate(delegate StreamNamer) {
	r.streamNameDelegate = delegate
}

//Load loads the aggragate with the specified type and ID
func (r *CommonDomainRepository) Load(aggregateType string, id string) (AggregateRoot, error) {

	if r.aggregateFactory == nil {
		return nil, fmt.Errorf("The common domain repository has no Aggregate Factory.")
	}

	if r.streamNameDelegate == nil {
		return nil, fmt.Errorf("The common domain repository has no stream name delegate.")
	}

	if r.eventFactory == nil {
		return nil, fmt.Errorf("The common domain has no Event Factory.")
	}

	aggregate := r.aggregateFactory.GetAggregate(aggregateType, id)
	if aggregate == nil {
		return nil, fmt.Errorf("The repository has no aggregate factory registered for aggregate type: %s", aggregateType)
	}

	stream, err := r.streamNameDelegate.GetStreamName(aggregateType, id)
	if err != nil {
		return nil, err
	}

	eventsChannel := r.eventStore.ReadStreamForwardAsync(stream, nil, nil, 0)

	for {
		select {
		case ev, open := <-eventsChannel:
			if !open {
				return aggregate, nil
			}

			if ev.Err != nil {
				if ev.Resp.StatusCode == http.StatusNotFound {
					return nil, &AggregateNotFoundError{AggregateType: aggregateType, AggregateID: id}
				}

				return nil, ev.Err
			}

			event := r.eventFactory.GetEvent(ev.EventResp.Event.EventType)
			if event == nil {
				return nil, fmt.Errorf("The event type %s is not registered with the eventstore.", ev.EventResp.Event.EventType)
			}

			data, ok := ev.EventResp.Event.Data.(*json.RawMessage)
			if !ok {
				return nil, fmt.Errorf("Common domain repository could not unmarshal even. Event data is not of type *json.RawMessage")
			}

			if err := json.Unmarshal(*data, event); err != nil {
				return nil, err
			}
			em := NewEventMessage(id, event)
			aggregate.Apply(em, false)
			aggregate.IncrementVersion()
		}
	}
}

//Save persists an aggregate
func (r *CommonDomainRepository) Save(aggregate AggregateRoot) error {

	if r.streamNameDelegate == nil {
		return fmt.Errorf("The common domain repository has no stream name delagate.")
	}

	resultEvents := aggregate.GetChanges()

	expectedVersion := aggregate.Version() - 1

	streamName, err := r.streamNameDelegate.GetStreamName(typeOf(aggregate), aggregate.AggregateID())
	if err != nil {
		return err
	}

	if len(resultEvents) > 0 {

		evs := make([]*goes.Event, len(resultEvents))

		for k, v := range resultEvents {
			//TODO: There is no test for this code
			v.SetHeader("AggregateID", aggregate.AggregateID())
			evs[k] = goes.ToEventData("", v.EventType(), v.Event(), v.GetHeaders())
		}

		resp, err := r.eventStore.AppendToStream(streamName, &goes.StreamVersion{Number: expectedVersion}, evs...)
		if err != nil {
			if resp.StatusCode == http.StatusBadRequest {
				return &ConcurrencyError{Aggregate: aggregate, ExpectedVersion: expectedVersion, StreamName: streamName}
			}

			return err
		}
	}

	aggregate.ClearChanges()

	//TODO: Write tests to verify this
	for _, v := range resultEvents {
		r.eventBus.PublishEvent(v)
	}

	return nil
}
