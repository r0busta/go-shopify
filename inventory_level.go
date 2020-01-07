package goshopify

import (
	"fmt"
)

const inventoryLevelsBasePath = "inventory_levels"

// InventoryLevelService is an interface for interacting with the
// inventory levels endpoints of the Shopify API
// See https://help.shopify.com/en/api/reference/inventory/inventorylevel
type InventoryLevelService interface {
	Set(InventoryLevel) (*InventoryLevel, error)
}

// InventoryLevelServiceOp is the default implementation of the InventoryLevelService interface
type InventoryLevelServiceOp struct {
	client *Client
}

// InventoryLevel represents a Shopify inventory item
type InventoryLevel struct {
	InventoryItemID int64 `json:"inventory_item_id,omitempty"`
	LocationID      int64 `json:"location_id,omitempty"`
	Available       int   `json:"available"`
}

// InventoryLevelResource is used for handling single item requests and responses
type InventoryLevelResource struct {
	InventoryLevel *InventoryLevel `json:"inventory_level"`
}

// InventoryLevelsResource is used for handling multiple item responsees
type InventoryLevelsResource struct {
	InventoryLevels []InventoryLevel `json:"inventory_levels"`
}

// Set inventory level
func (s *InventoryLevelServiceOp) Set(level InventoryLevel) (*InventoryLevel, error) {
	path := fmt.Sprintf("%s/%s/%s.json", globalApiPathPrefix, inventoryLevelsBasePath, "set")
	resource := new(InventoryLevelResource)
	err := s.client.Post(path, level, resource)
	return resource.InventoryLevel, err
}
