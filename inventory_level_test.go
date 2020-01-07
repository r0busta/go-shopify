package goshopify

import (
	"testing"
)

func inventoryLevelTests(t *testing.T, level *InventoryLevel) {
	if level == nil {
		t.Errorf("InventoryLevel is nil")
		return
	}

	expectedInt := int64(808950810)
	if level.InventoryItemID != expectedInt {
		t.Errorf("InventoryLevel.InventoryItemID returned %+v, expected %+v", level.InventoryItemID, expectedInt)
	}
}

func inventoryLevelsTests(t *testing.T, levels []InventoryLevel) {
	expectedLen := 3
	if len(levels) != expectedLen {
		t.Errorf("InventoryLevels list length is %+v, expected %+v", len(levels), expectedLen)
	}
}
