package service_test

import (
	"testing"

	"myapp/internal"
	"myapp/internal/service"
	"myapp/internal/testutil"
)

func TestCreateOrderDecreasesStock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := service.NewOrderService(db)

	product := internal.Product{Name: "Test Widget", SKU: "TST-001", Price: 25.0}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	warehouse := internal.Warehouse{Name: "Main", Location: "HQ"}
	if err := db.Create(&warehouse).Error; err != nil {
		t.Fatalf("create warehouse: %v", err)
	}

	inv := internal.Inventory{ProductID: product.ID, WarehouseID: warehouse.ID, Quantity: 100}
	if err := db.Create(&inv).Error; err != nil {
		t.Fatalf("create inventory: %v", err)
	}

	order, err := svc.CreateOrder(service.CreateOrderInput{
		CustomerName: "Acme Corp",
		Items: []service.OrderLineInput{{
			ProductID: product.ID, WarehouseID: warehouse.ID, Quantity: 10,
		}},
		Actor: "test@ims.local",
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if order == nil || order.TotalAmount != 250 {
		t.Fatalf("unexpected order: %+v", order)
	}

	qty, err := svc.GetInventoryQty(product.ID, warehouse.ID)
	if err != nil {
		t.Fatalf("GetInventoryQty: %v", err)
	}
	if qty != 90 {
		t.Fatalf("expected stock 90, got %d", qty)
	}

	var movements int64
	db.Model(&internal.StockMovement{}).Where("product_id = ? AND type = ?", product.ID, "OUT").Count(&movements)
	if movements != 1 {
		t.Fatalf("expected 1 OUT movement, got %d", movements)
	}
}

func TestCreateOrderInsufficientStock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := service.NewOrderService(db)

	product := internal.Product{Name: "Low Stock", SKU: "LOW-001", Price: 10}
	db.Create(&product)
	warehouse := internal.Warehouse{Name: "WH1"}
	db.Create(&warehouse)
	db.Create(&internal.Inventory{ProductID: product.ID, WarehouseID: warehouse.ID, Quantity: 2})

	_, err := svc.CreateOrder(service.CreateOrderInput{
		CustomerName: "Buyer",
		Items: []service.OrderLineInput{{
			ProductID: product.ID, WarehouseID: warehouse.ID, Quantity: 5,
		}},
		Actor: "test@ims.local",
	})
	if err == nil {
		t.Fatal("expected insufficient stock error")
	}

	qty, _ := svc.GetInventoryQty(product.ID, warehouse.ID)
	if qty != 2 {
		t.Fatalf("stock should remain 2, got %d", qty)
	}
}

func TestCreateOrderRollsBackOnFailure(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := service.NewOrderService(db)

	product := internal.Product{Name: "P1", SKU: "P1", Price: 10}
	db.Create(&product)
	warehouse := internal.Warehouse{Name: "WH"}
	db.Create(&warehouse)
	db.Create(&internal.Inventory{ProductID: product.ID, WarehouseID: warehouse.ID, Quantity: 100})

	_, err := svc.CreateOrder(service.CreateOrderInput{
		CustomerName: "Buyer",
		Items: []service.OrderLineInput{
			{ProductID: product.ID, WarehouseID: warehouse.ID, Quantity: 5},
			{ProductID: 9999, WarehouseID: warehouse.ID, Quantity: 1},
		},
		Actor: "test@ims.local",
	})
	if err == nil {
		t.Fatal("expected error for missing product")
	}

	var orderCount int64
	db.Model(&internal.Order{}).Count(&orderCount)
	if orderCount != 0 {
		t.Fatalf("order should not be created on failure, got %d orders", orderCount)
	}

	qty, _ := svc.GetInventoryQty(product.ID, warehouse.ID)
	if qty != 100 {
		t.Fatalf("stock should remain 100 after rollback, got %d", qty)
	}
}
