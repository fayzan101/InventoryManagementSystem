package testutil

import (
	"testing"

	"myapp/internal"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	internal.DB = db

	if err := db.AutoMigrate(
		&internal.Category{},
		&internal.Product{},
		&internal.Warehouse{},
		&internal.Inventory{},
		&internal.StockMovement{},
		&internal.Supplier{},
		&internal.Customer{},
		&internal.PurchaseOrder{},
		&internal.POItem{},
		&internal.Order{},
		&internal.OrderItem{},
		&internal.StockTransfer{},
		&internal.StockTransferItem{},
		&internal.Employee{},
		&internal.User{},
		&internal.AuditLog{},
	); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}
