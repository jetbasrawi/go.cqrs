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

// DomainRepository is the interface that all domain repositories should implement.
type DomainRepository interface {
	//Loads an aggregate of the given type and ID
	Load(aggregateTypeName string, aggregateID string) (AggregateRoot, error)

	//Saves the aggregate.
	Save(aggregate AggregateRoot, expectedVersion *int) error
}

// GetEventStoreCommonDomainRepo is an implementation of the DomainRepository
// that uses GetEventStore for persistence
type GetEventStoreCommonDomainRepo struct {
	eventStore         *goes.Client
	eventBus           EventBus
	streamNameDelegate StreamNamer
	aggregateFactory   AggregateFactory
	eventFactory       EventFactory
}

// NewCommonDomainRepository constructs a new CommonDomainRepository
func NewCommonDomainRepository(eventStore *goes.Client, eventBus EventBus) (*GetEventStoreCommonDomainRepo, error) {
	if eventStore == nil {
		return nil, fmt.Errorf("Nil Eventstore injected into repository.")
	}

	if eventBus == nil {
		return nil, fmt.Errorf("Nil EventBus injected into repository.")
	}

	d := &GetEventStoreCommonDomainRepo{
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
func (r *GetEventStoreCommonDomainRepo) SetAggregateFactory(factory AggregateFactory) {
	r.aggregateFactory = factory
}

// SetEventFactory sets the event factory that should be used to instantiate event
// instances.
//
// Only one event factory can be set at a time. Any subsequent registration will
// overwrite the previous factory.
func (r *GetEventStoreCommonDomainRepo) SetEventFactory(factory EventFactory) {
	r.eventFactory = factory
}

// SetStreamNameDelegate sets the stream name delegate
func (r *GetEventStoreCommonDomainRepo) SetStreamNameDelegate(delegate StreamNamer) {
	r.streamNameDelegate = delegate
}

// Load will load all events from a stream and apply those events to an aggregate
// of the type specified.
//
// The aggregate type and id will be passed to the configured StreamNamer to
// get the stream name.
func (r *GetEventStoreCommonDomainRepo) Load(aggregateType, id string) (AggregateRoot, error) {

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
			return nil, &ErrRepositoryUnavailable{}
		case *goes.ErrNoMoreEvents:
			return aggregate, nil
		case *goes.ErrUnauthorized:
			return nil, &ErrUnauthorized{}
		case *goes.ErrNotFound:
			return nil, &ErrAggregateNotFound{AggregateType: aggregateType, AggregateID: id}
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
		em := NewEventMessage(id, event, Int(stream.EventResponse().Event.EventNumber))
		for k, v := range meta {
			em.SetHeader(k, v)
		}
		aggregate.Apply(em, false)
		aggregate.IncrementVersion()
	}

	return aggregate, nil

}

// Save persists an aggregate
func (r *GetEventStoreCommonDomainRepo) Save(aggregate AggregateRoot, expectedVersion *int) error {

	if r.streamNameDelegate == nil {
		return fmt.Errorf("The common domain repository has no stream name delagate.")
	}

	resultEvents := aggregate.GetChanges()

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
		err := streamWriter.Append(expectedVersion, evs...)
		switch e := err.(type) {
		case nil:
			break
		case *goes.ErrConcurrencyViolation:
			return &ErrConcurrencyViolation{Aggregate: aggregate, ExpectedVersion: expectedVersion, StreamName: streamName}
		case *goes.ErrUnauthorized:
			return &ErrUnauthorized{}
		case *goes.ErrTemporarilyUnavailable:
			return &ErrRepositoryUnavailable{}
		default:
			return &ErrUnexpected{Err: e}
		}
	}

	aggregate.ClearChanges()

	for k, v := range resultEvents {
		if expectedVersion == nil {
			r.eventBus.PublishEvent(v)
		} else {
			em := NewEventMessage(v.AggregateID(), v.Event(), Int(*expectedVersion+k+1))
			r.eventBus.PublishEvent(em)
		}
	}

	return nil
}
