// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"fmt"
	"net/url"

	"github.com/jetbasrawi/go.geteventstore"
)

// Error is returned when the eventstore is temporarily unavailable
type RepositoryUnavailableError struct{}

func (e *RepositoryUnavailableError) Error() string {
	return "The repository is temporarily unavailable."
}

// AggregateNotFoundError error returned when an aggregate was not found in the repository.
type AggregateNotFoundError struct {
	AggregateID   string
	AggregateType string
}

func (e *AggregateNotFoundError) Error() string {
	return fmt.Sprintf("Could not find any aggregate of type %s with id %s",
		e.AggregateType,
		e.AggregateID)
}

// DomainRepository is the interface that all domain repositories should implement.
type DomainRepository interface {
	//Loads an aggregate of the given type and ID
	Load(string, string) (AggregateRoot, error)

	//Saves the aggregate.
	Save(AggregateRoot) error
}

// CommonDomainRepository is a generic repository implementation
type CommonDomainRepository struct {
	eventStore         *goes.Client
	eventBus           EventBus
	streamNameDelegate StreamNamer
	aggregateFactory   AggregateFactory
	eventFactory       EventFactory
}

// NewCommonDomainRepository constructs a new CommonDomainRepository
func NewCommonDomainRepository(eventStore *goes.Client, eventBus EventBus) (*CommonDomainRepository, error) {
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

// SetAggregateFactory sets the aggregate factory that should be used to
// instantate aggregate instances
//
// Only one AggregateFactory can be registered at any one time.
// Any registration will overwrite the provious registration.
func (r *CommonDomainRepository) SetAggregateFactory(factory AggregateFactory) {
	r.aggregateFactory = factory
}

// SetEventFactory sets the event factory that should be used to instantiate event
// instances.
//
// Only one event factory can be set at a time. Any subsequent registration will
// overwrite the previous factory.
func (r *CommonDomainRepository) SetEventFactory(factory EventFactory) {
	r.eventFactory = factory
}

// SetStreamNameDelegate sets the stream name delegate
func (r *CommonDomainRepository) SetStreamNameDelegate(delegate StreamNamer) {
	r.streamNameDelegate = delegate
}

// Load will load all events from a stream and apply those events to an aggregate
// of the type specified.
//
// The aggregate type and id will be passed to the configured StreamNamer to
// get the stream name.
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

	streamName, err := r.streamNameDelegate.GetStreamName(aggregateType, id)
	if err != nil {
		return nil, err
	}

	stream := r.eventStore.NewStreamReader(streamName)
	for stream.Next() {
		switch err := stream.Err().(type) {
		case nil:
			break
		case *url.Error, *goes.ErrTemporarilyUnavailable:
			return nil, &RepositoryUnavailableError{}
		case *goes.ErrNoMoreEvents:
			return aggregate, nil
		case *goes.ErrUnauthorized:
			return nil, &ErrUnauthorized{}
		case *goes.ErrNotFound:
			return nil, &AggregateNotFoundError{AggregateType: aggregateType, AggregateID: id}
		default:
			return nil, &ErrUnexpected{Err: err}
		}

		event := r.eventFactory.GetEvent(stream.EventResponse().Event.EventType)

		//TODO: No test for meta
		meta := make(map[string]string)
		stream.Scan(event, &meta)
		if stream.Err() != nil {
			return nil, stream.Err()
		}
		em := NewEventMessage(id, event)
		for k, v := range meta {
			em.SetHeader(k, v)
		}
		aggregate.Apply(em, false)
		aggregate.IncrementVersion()
	}

	return aggregate, nil

}

// Save persists an aggregate
func (r *CommonDomainRepository) Save(aggregate AggregateRoot) error {

	if r.streamNameDelegate == nil {
		return fmt.Errorf("The common domain repository has no stream name delagate.")
	}

	resultEvents := aggregate.GetChanges()

	expectedVersion := aggregate.Version()

	streamName, err := r.streamNameDelegate.GetStreamName(typeOf(aggregate), aggregate.AggregateID())
	if err != nil {
		return err
	}

	if len(resultEvents) > 0 {

		evs := make([]*goes.Event, len(resultEvents))

		for k, v := range resultEvents {
			//TODO: There is no test for this code
			v.SetHeader("AggregateID", aggregate.AggregateID())
			evs[k] = goes.NewEvent("", v.EventType(), v.Event(), v.GetHeaders())
		}

		streamWriter := r.eventStore.NewStreamWriter(streamName)
		err := streamWriter.Append(&expectedVersion, evs...)
		switch e := err.(type) {
		case nil:
			break
		case *goes.ErrConcurrencyViolation:
			return &ConcurrencyError{Aggregate: aggregate, ExpectedVersion: expectedVersion, StreamName: streamName}
		case *goes.ErrUnauthorized:
			return &ErrUnauthorized{}
		case *goes.ErrTemporarilyUnavailable:
			return &RepositoryUnavailableError{}
		default:
			return &ErrUnexpected{Err: e}
		}
	}

	aggregate.ClearChanges()

	//TODO: Write tests to verify this
	for _, v := range resultEvents {
		r.eventBus.PublishEvent(v)
	}

	return nil
}
