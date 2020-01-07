package goshopify

import (
	"fmt"
	"testing"

	"gopkg.in/jarcoal/httpmock.v1"
)

func inventoryLevelTests(t *testing.T, level *InventoryLevel) {
	if level == nil {
		t.Errorf("InventoryLevel is nil")
		return
	}

	expectedInt := int64(808950810)
	if level.ID != expectedInt {
		t.Errorf("InventoryLevel.ID returned %+v, expected %+v", level.ID, expectedInt)
	}

	expectedSKU := "new sku"
	if level.SKU != expectedSKU {
		t.Errorf("InventoryLevel.SKU sku is %+v, expected %+v", level.SKU, expectedSKU)
	}

	if level.Cost == nil {
		t.Errorf("InventoryLevel.Cost is nil")
		return
	}

	expectedCost := 25.00
	costFloat, _ := level.Cost.Float64()
	if costFloat != expectedCost {
		t.Errorf("InventoryLevel.Cost (float) is %+v, expected %+v", costFloat, expectedCost)
	}
}

func inventoryLevelsTests(t *testing.T, levels []InventoryLevel) {
	expectedLen := 3
	if len(levels) != expectedLen {
		t.Errorf("InventoryLevels list lenth is %+v, expected %+v", len(levels), expectedLen)
	}
}

func TestInventoryLevelsList(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://fooshop.myshopify.com/%s/inventory_levels.json", globalApiPathPrefix),
		httpmock.NewBytesResponder(200, loadFixture("inventory_levels.json")))

	levels, err := client.InventoryLevel.List(nil)
	if err != nil {
		t.Errorf("InventoryLevels.List returned error: %v", err)
	}

	inventoryLevelsTests(t, levels)
}

func TestInventoryLevelGet(t *testing.T) {
	setup()
	defer teardown()

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://fooshop.myshopify.com/%s/inventory_levels/set.json", globalApiPathPrefix),
		httpmock.NewBytesResponder(200, loadFixture("inventory_level.json")))

	level, err := client.InventoryLevel.Get(1, nil)
	if err != nil {
		t.Errorf("InventoryLevel.Get returned error: %v", err)
	}

	inventoryLevelTests(t, level)
}
