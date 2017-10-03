package main

import (
	"log"
	"net/http"

	"github.com/jetbasrawi/go.cqrs"
	"github.com/jetbasrawi/go.cqrs/examples/simplecqrs/simplecqrs"
	"github.com/jetbasrawi/go.geteventstore"
)

var (
	readModel  simplecqrs.ReadModelFacade
	dispatcher ycq.Dispatcher
)

func init() {

	// CQRS Infrastructure configuration

	// Configure the read model

	// Create a readModel instance
	readModel = simplecqrs.NewReadModel()

	// Create a InventoryListView
	listView := simplecqrs.NewInventoryListView()
	// Create a InventoryItemDetailView
	detailView := simplecqrs.NewInventoryItemDetailView()

	// Create an EventBus
	eventBus := ycq.NewInternalEventBus()
	// Register the listView as an event handler on the event bus
	// for the events specified.
	eventBus.AddHandler(listView,
		&simplecqrs.InventoryItemCreated{},
		&simplecqrs.InventoryItemRenamed{},
		&simplecqrs.InventoryItemDeactivated{},
	)
	// Register the detail view as an event handler on the event bus
	// for the events specified.
	eventBus.AddHandler(detailView,
		&simplecqrs.InventoryItemCreated{},
		&simplecqrs.InventoryItemRenamed{},
		&simplecqrs.InventoryItemDeactivated{},
		&simplecqrs.ItemsRemovedFromInventory{},
		&simplecqrs.ItemsCheckedIntoInventory{},
	)

	// Create an in memory repository
	//repo := simplecqrs.NewInMemoryRepo(eventBus)
	client, err := goes.NewClient(nil, "http://localhost:2113")
	if err != nil {
		log.Fatal(err)
	}
	repo, err := simplecqrs.NewInventoryItemRepo(client, eventBus)
	if err != nil {
		log.Fatal(err)
	}

	// Create an InventoryCommandHandlers instance
	inventoryCommandHandler := simplecqrs.NewInventoryCommandHandlers(repo)

	// Create a dispatcher
	dispatcher = ycq.NewInMemoryDispatcher()
	// Register the inventory command handlers instance as a command handler
	// for the events specified.
	dispatcher.RegisterHandler(inventoryCommandHandler,
		&simplecqrs.CreateInventoryItem{},
		&simplecqrs.DeactivateInventoryItem{},
		&simplecqrs.RenameInventoryItem{},
		&simplecqrs.CheckInItemsToInventory{},
		&simplecqrs.RemoveItemsFromInventory{},
	)

}

func main() {

	mux := setupHandlers()
	if err := http.ListenAndServe(":8088", mux); err != nil {
		log.Fatal(err)
	}

}
