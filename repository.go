package ycq

import (
	"encoding/json"
	"fmt"

	"github.com/jetbasrawi/goes"
	"github.com/jetbasrawi/yoono-uuid"
)

type DomainRepository interface {
	Load(string, uuid.UUID) (AggregateRoot, error)

	Save(AggregateRoot) error
}

type CommonDomainRepository struct {
	eventStore         GetEventStoreRepositoryClient
	streamNameDelegate StreamNamer
	aggregateFactory   AggregateFactory
	eventFactory       EventFactory
}

func NewCommonDomainRepository(eventStore GetEventStoreRepositoryClient) (*CommonDomainRepository, error) {
	if eventStore == nil {
		return nil, fmt.Errorf("Nil Eventstore injected into repository.")
	}
	d := &CommonDomainRepository{
		eventStore: eventStore,
	}
	return d, nil
}

func (r *CommonDomainRepository) SetAggregateFactory(factory AggregateFactory) {
	r.aggregateFactory = factory
}

func (r *CommonDomainRepository) SetEventFactory(factory EventFactory) {
	r.eventFactory = factory
}

func (r *CommonDomainRepository) SetStreamNameDelegate(delegate StreamNamer) {
	r.streamNameDelegate = delegate
}

func (r *CommonDomainRepository) Load(aggregateType string, id uuid.UUID) (AggregateRoot, error) {

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

	eventsChannel := r.eventStore.ReadStreamForwardAsync(stream, nil, nil)

	for {
		select {
		case ev, open := <-eventsChannel:
			if !open {
				return aggregate, nil
			}

			if ev.Err != nil {
				//TODO
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
			aggregate.Apply(em)
			aggregate.IncrementVersion()

		}
	}
}

// Save  persists an aggregate
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
			evs[k] = goes.ToEventData("", v.EventType(), v.Event(), v.GetHeaders())
		}

		resp, err := r.eventStore.AppendToStream(streamName, &goes.StreamVersion{expectedVersion}, evs...)
		if err != nil {
			return fmt.Errorf("%s", resp.StatusMessage)
		}
	}

	aggregate.ClearChanges()

	return nil
}
