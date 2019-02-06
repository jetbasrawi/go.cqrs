package simplecqrs

import (
	"github.com/jetbasrawi/go.cqrs"
)

// InMemoryRepo provides an in memory repository implementation.
type InMemoryRepo struct {
	current   map[string][]ycq.EventMessage
	publisher ycq.EventBus
}

// NewInMemoryRepo constructs an InMemoryRepo instance.
func NewInMemoryRepo(eventBus ycq.EventBus) *InMemoryRepo {
	return &InMemoryRepo{
		current:   make(map[string][]ycq.EventMessage),
		publisher: eventBus,
	}
}

// Load loads an aggregate of the specified type.
func (r *InMemoryRepo) Load(aggregateType, id string) (*InventoryItem, error) {

	events, ok := r.current[id]
	if !ok {
		return nil, &ycq.ErrAggregateNotFound{}
	}

	inventoryItem := NewInventoryItem(id)

	for _, v := range events {
		inventoryItem.Apply(v, false)
		inventoryItem.IncrementVersion()
	}

	return inventoryItem, nil
}

// Save persists an aggregate.
func (r *InMemoryRepo) Save(aggregate ycq.AggregateRoot, _ *int) error {

	//TODO: Look at the expected version
	for _, v := range aggregate.GetChanges() {
		r.current[aggregate.AggregateID()] = append(r.current[aggregate.AggregateID()], v)
		r.publisher.PublishEvent(v)
	}

	return nil

}
