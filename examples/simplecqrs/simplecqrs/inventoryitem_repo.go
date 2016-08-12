package simplecqrs

import (
	"fmt"
	"reflect"

	"github.com/jetbasrawi/go.cqrs"
	"github.com/jetbasrawi/go.geteventstore"
)

// InventoryItemRepo is a repository specialized for persistence of
// InventoryItems.
//
// While it is not required to construct a repository specialized for a
// specific aggregate type, it is better to do so. There can be quite a lot of
// repository configuration that is specific to a type and it is cleaner if that
// code is contained in a specialized repository as shown here.
// Also because the CommonDomainRepository Load method returns an interface{}, a
// type assertion is required. Here the type assertion is contained in this specialized
// repo and a *InventoryItem is returned from the repo.
type InventoryItemRepo struct {
	repo *ycq.GetEventStoreCommonDomainRepo
}

// NewInventoryItemRepo constructs a new InventoryItemRepository.
func NewInventoryItemRepo(eventStore *goes.Client, eventBus ycq.EventBus) (*InventoryItemRepo, error) {

	r, err := ycq.NewCommonDomainRepository(eventStore, eventBus)
	if err != nil {
		return nil, err
	}

	ret := &InventoryItemRepo{
		repo: r,
	}

	// An aggregate factory creates an aggregate instance given the name of an aggregate.
	aggregateFactory := ycq.NewDelegateAggregateFactory()
	aggregateFactory.RegisterDelegate(&InventoryItem{},
		func(id string) ycq.AggregateRoot { return NewInventoryItem(id) })
	ret.repo.SetAggregateFactory(aggregateFactory)

	// A stream name delegate constructs a stream name.
	// A common way to construct a stream name is to use a bounded context and
	// an aggregate id.
	// The interface for a stream name delegate takes a two strings. One may be
	// the aggregate type and the other the aggregate id. In this case the first
	// argument and the second argument are concatenated with a hyphen.
	streamNameDelegate := ycq.NewDelegateStreamNamer()
	streamNameDelegate.RegisterDelegate(func(t string, id string) string {
		return t + "-" + id
	}, &InventoryItem{})
	ret.repo.SetStreamNameDelegate(streamNameDelegate)

	// An event factory creates an instance of an event given the name of an event
	// as a string.
	eventFactory := ycq.NewDelegateEventFactory()
	eventFactory.RegisterDelegate(&InventoryItemCreated{},
		func() interface{} { return &InventoryItemCreated{} })
	eventFactory.RegisterDelegate(&InventoryItemRenamed{},
		func() interface{} { return &InventoryItemRenamed{} })
	eventFactory.RegisterDelegate(&InventoryItemDeactivated{},
		func() interface{} { return &InventoryItemDeactivated{} })
	eventFactory.RegisterDelegate(&ItemsRemovedFromInventory{},
		func() interface{} { return &ItemsRemovedFromInventory{} })
	eventFactory.RegisterDelegate(&ItemsCheckedIntoInventory{},
		func() interface{} { return &ItemsCheckedIntoInventory{} })
	ret.repo.SetEventFactory(eventFactory)

	return ret, nil
}

// Load loads events for an aggregate.
//
// Returns an *InventoryAggregate.
func (r *InventoryItemRepo) Load(aggregateType, id string) (*InventoryItem, error) {
	ar, err := r.repo.Load(reflect.TypeOf(&InventoryItem{}).Elem().Name(), id)
	if _, ok := err.(*ycq.ErrAggregateNotFound); ok {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if ret, ok := ar.(*InventoryItem); ok {
		return ret, nil
	}

	return nil, fmt.Errorf("Could not cast aggregate returned to type of %s", reflect.TypeOf(&InventoryItem{}).Elem().Name())
}

// Save persists an aggregate.
func (r *InventoryItemRepo) Save(aggregate ycq.AggregateRoot, expectedVersion *int) error {
	return r.repo.Save(aggregate, expectedVersion)
}
