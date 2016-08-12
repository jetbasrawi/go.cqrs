package simplecqrs

// CreateInventoryItem create a new inventory item
type CreateInventoryItem struct {
	Name string
}

// DeactivateInventoryItem deactivates the inventory item
type DeactivateInventoryItem struct {
	OriginalVersion int
}

// RenameInventoryItem renames an inventory item
type RenameInventoryItem struct {
	OriginalVersion int
	NewName         string
}

// CheckInItemsToInventory adds items to inventory
type CheckInItemsToInventory struct {
	OriginalVersion int
	Count           int
}

// RemoveItemsFromInventory removes items from inventory
type RemoveItemsFromInventory struct {
	OriginalVersion int
	Count           int
}
