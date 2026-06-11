package internal

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(connStr string) {
	var err error
	DB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to PostgreSQL database with GORM.")
	if err := DB.AutoMigrate(
		&Category{},
		&Product{},
		&Warehouse{},
		&Inventory{},
		&StockMovement{},
		&Supplier{},
		&Customer{},
		&PurchaseOrder{},
		&POItem{},
		&Order{},
		&OrderItem{},
		&StockTransfer{},
		&StockTransferItem{},
		&Employee{},
		&AuditLog{},
	); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}
	log.Println("Database migration completed successfully.")
}
func LogAudit(action, entity string, entityID uint, userID, details string) {
	log := AuditLog{
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		UserID:    userID,
		Details:   details,
		IPAddress: "127.0.0.1",
		CreatedAt: time.Now(),
	}
	DB.Create(&log)
}
