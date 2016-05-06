package ycq

import (
	"fmt"
	"github.com/jetbasrawi/yoono-uuid"
)

type DomainRepository interface {
	Load(string, uuid.UUID) (AggregateRoot, error)

	Save(AggregateRoot) error
}

type CommonDomainRepository struct {
	eventStore         EventStore
	streamNameDelegate StreamNamer
	aggregateFactory   AggregateFactory
}

func NewCommonDomainRepository(eventStore EventStore) (*CommonDomainRepository, error) {
	if eventStore == nil {
		return nil, fmt.Errorf("Nil Eventstore injected into repository.")
	}
	d := &CommonDomainRepository{
		eventStore: eventStore,
	}
	return d, nil
}

func (r *CommonDomainRepository) RegisterAggregateFactory(factory AggregateFactory) error {
	r.aggregateFactory = factory
	return nil
}

func (r *CommonDomainRepository) Load(aggregateType string, id uuid.UUID) (AggregateRoot, error) {

	if r.aggregateFactory == nil {
		return nil, fmt.Errorf("The common domain repository has no Aggregate Factory.")
	}

	if r.streamNameDelegate == nil {
		return nil, fmt.Errorf("The common domain repository has no stream name delegate.")
	}

	aggregate := r.aggregateFactory.GetAggregate(aggregateType, id)
	if aggregate == nil {
		return nil, fmt.Errorf("The repository has no aggregate factory registered for aggregate type: %s", aggregateType)
	}

	stream, err := r.streamNameDelegate.GetStreamName(aggregateType, id)
	if err != nil {
		return nil, err
	}

	events, err := r.eventStore.Load(stream)
	if err == ErrNoEventsFound {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	for _, event := range events {
		aggregate.Apply(event)
		aggregate.IncrementVersion()
	}

	return aggregate, nil
}

// Save  persists an aggregate
func (r *CommonDomainRepository) Save(aggregate AggregateRoot) error {

	if r.streamNameDelegate == nil {
		return fmt.Errorf("The common domain repository has no stream name delagate.")
	}

	resultEvents := aggregate.GetChanges()

	expectedVersion := aggregate.Version() - 1

	streamName, err := r.streamNameDelegate.GetStreamName(aggregate.AggregateType(), aggregate.AggregateID())
	if err != nil {
		return err
	}

	if len(resultEvents) > 0 {
		err := r.eventStore.Save(streamName, resultEvents, &expectedVersion, nil)
		if err != nil {
			return err
		}
	}

	aggregate.ClearChanges()

	return nil
}
