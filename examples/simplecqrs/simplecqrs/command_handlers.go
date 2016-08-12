package simplecqrs

import (
	"log"
	"reflect"

	"github.com/jetbasrawi/go.cqrs"
)

type InventoryItemRepository interface {
	Load(string, string) (*InventoryItem, error)
	Save(ycq.AggregateRoot, *int) error
}

// InventoryCommandHandlers provides methods for processing commands related
// to inventory items.
type InventoryCommandHandlers struct {
	repo InventoryItemRepository
}

// NewInventoryCommandHandlers contructs a new InventoryCommandHandlers
func NewInventoryCommandHandlers(repo InventoryItemRepository) *InventoryCommandHandlers {
	return &InventoryCommandHandlers{
		repo: repo,
	}
}

// Handle processes inventory item commands.
func (h *InventoryCommandHandlers) Handle(message ycq.CommandMessage) error {

	var item *InventoryItem

	switch cmd := message.Command().(type) {

	case *CreateInventoryItem:

		item = NewInventoryItem(message.AggregateID())
		if err := item.Create(cmd.Name); err != nil {
			return &ycq.ErrCommandExecution{Command: message, Reason: err.Error()}
		}
		return h.repo.Save(item, ycq.Int(item.OriginalVersion()))

	case *DeactivateInventoryItem:

		item, _ = h.repo.Load(reflect.TypeOf(&InventoryItem{}).Elem().Name(), message.AggregateID())
		if err := item.Deactivate(); err != nil {
			return &ycq.ErrCommandExecution{Command: message, Reason: err.Error()}
		}
		return h.repo.Save(item, ycq.Int(item.OriginalVersion()))

	case *RemoveItemsFromInventory:

		item, _ = h.repo.Load(reflect.TypeOf(&InventoryItem{}).Elem().Name(), message.AggregateID())
		item.Remove(cmd.Count)
		return h.repo.Save(item, ycq.Int(item.OriginalVersion()))

	case *CheckInItemsToInventory:

		item, _ = h.repo.Load(reflect.TypeOf(&InventoryItem{}).Elem().Name(), message.AggregateID())
		item.CheckIn(cmd.Count)
		return h.repo.Save(item, ycq.Int(item.OriginalVersion()))

	case *RenameInventoryItem:

		item, _ = h.repo.Load(reflect.TypeOf(&InventoryItem{}).Elem().Name(), message.AggregateID())
		if err := item.ChangeName(cmd.NewName); err != nil {
			return &ycq.ErrCommandExecution{Command: message, Reason: err.Error()}
		}
		return h.repo.Save(item, ycq.Int(item.OriginalVersion()))

	default:
		log.Fatalf("InventoryCommandHandlers has received a command that it is does not know how to handle, %#v", cmd)
	}

	return nil
}
