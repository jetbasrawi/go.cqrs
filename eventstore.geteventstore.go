package ycq

import (
	"encoding/json"
	"fmt"
	"github.com/jetbasrawi/goes"
	"github.com/jetbasrawi/yoono-uuid"
	"net/http"
)

type GetEventStore struct {
	eventBus     EventBus
	client       *goes.Client
	eventFactory EventFactory
}

func NewGetEventStore(eventBus EventBus, eventstoreURL string) (*GetEventStore, error) {
	goesClient, err := goes.NewClient(nil, eventstoreURL)
	if err != nil {
		return nil, err
	}

	s := &GetEventStore{
		eventBus: eventBus,
		client:   goesClient,
	}

	return s, nil
}

func (s *GetEventStore) SetEventFactory(eventFactory EventFactory) {
	s.eventFactory = eventFactory
}

// Load loads all events for the aggregate id from the memory store.
// Returns ErrNoEventsFound if no events can be found.
func (s *GetEventStore) Load(stream string) ([]EventMessage, error) {

	events, _, err := s.client.ReadFeedForward(stream, nil, nil)
	if err != nil {
		if e, ok := err.(*goes.ErrorResponse); ok {
			if e.StatusCode == http.StatusNotFound {
				return nil, ErrNoEventsFound
			}
		} else {
			return nil, err
		}
	}

	if len(events) <= 0 {
		return nil, ErrNoEventsFound
	}

	ret := make([]EventMessage, len(events))
	for i, r := range events {
		ev := s.eventFactory.GetEvent(r.Event.EventType)
		if ev == nil {
			return nil, fmt.Errorf("The event type %s is not registered with the eventstore.", r.Event.EventType)
		}

		if data, ok := r.Event.Data.(*json.RawMessage); ok {
			if err := json.Unmarshal(*data, ev); err != nil {
				return nil, err
			}
		}

		id, err := uuid.FromString(r.ID)
		if err != nil {
			return nil, err
		}

		ret[i] = NewEventMessage(id, ev)
	}

	return ret, nil
}

func (s *GetEventStore) Save(stream string, events []EventMessage, expectedVersion *int, headers map[string]interface{}) error {

	if len(events) == 0 {
		return ErrNoEventsToAppend
	}

	streamName := fmt.Sprintf("%s", events[0].AggregateID())
	esEvents := make([]*goes.Event, len(events))

	for k, v := range events {
		esEvents[k] = s.client.ToEventData("", v.EventType(), v, headers)
	}

	var version *goes.StreamVersion
	if expectedVersion != nil {
		version = &goes.StreamVersion{Number: *expectedVersion}
	}
	_, err := s.client.AppendToStream(streamName, version, esEvents...)
	if err != nil {
		return err //TODO: Much improvement
	}

	if s.eventBus != nil {
		for _, v := range events {
			s.eventBus.PublishEvent(v)
		}
	}

	return nil
}
