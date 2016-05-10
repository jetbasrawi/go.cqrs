package ycq

import (
	//"encoding/json"
	"fmt"

	"github.com/jetbasrawi/goes"
	"github.com/jetbasrawi/yoono-uuid"
)

type DomainRepository interface {
	Load(string, uuid.UUID) (AggregateRoot, error)

	Save(AggregateRoot) error
}

//type GESRepositoryClient struct {
//client *goes.Client
//}

//func (c *GESRepositoryClient) ReadStreamForwardAsync(stream string, version *goes.StreamVersion, take *goes.Take) <-chan struct {
//*goes.EventResponse
//*goes.Response
//error
//} {
//ch := c.client.ReadStreamForwardAsync(stream, version, take)
//return ch
//}

//func (c GESRepositoryClient) AppendToStream(stream string, version *goes.StreamVersion, events ...*goes.Event) (*goes.Response, error) {
//return c.client.AppendToStream(stream, version, events)
//}

type CommonDomainRepository struct {
	eventStore         goes.GetEventStoreRepositoryClient
	streamNameDelegate StreamNamer
	aggregateFactory   AggregateFactory
	eventFactory       EventFactory
}

func NewCommonDomainRepository(eventStore goes.GetEventStoreRepositoryClient) (*CommonDomainRepository, error) {
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

	aggregate := r.aggregateFactory.GetAggregate(aggregateType, id)
	if aggregate == nil {
		return nil, fmt.Errorf("The repository has no aggregate factory registered for aggregate type: %s", aggregateType)
	}

	_, err := r.streamNameDelegate.GetStreamName(aggregateType, id)
	if err != nil {
		return nil, err
	}

	//eventsChannel := r.eventStore.ReadStreamForwardAsync(stream, nil, nil)

	//for {
	//select {
	//case ev, open := <-eventsChannel:
	//if !open {
	//break
	//}

	//if ev.error != nil {
	////TODO
	//}

	//event := r.eventFactory.GetEvent(ev.EventResponse.Event.EventType)
	//if event == nil {
	//return nil, fmt.Errorf("The event type %s is not registered with the eventstore.", ev.EventResponse.Event.EventType)
	//}

	//if data, ok := ev.EventResponse.Event.Data.(*json.RawMessage); ok {
	//if err := json.Unmarshal(*data, event); err != nil {
	//return nil, err
	//}
	//}

	//em := NewEventMessage(id, event)
	//aggregate.Apply(em)
	//aggregate.IncrementVersion()

	//}
	//}

	return aggregate, nil
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
