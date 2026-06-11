package internal

import (
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
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
		&User{},
		&AuditLog{},
	); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}
	log.Println("Database migration completed successfully.")
	SeedDefaultAdmin()
}

func SeedDefaultAdmin() {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" {
		email = "admin@ims.local"
	}
	if password == "" {
		password = "admin123"
	}

	var count int64
	DB.Model(&User{}).Count(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Warning: failed to hash admin password: %v", err)
		return
	}

	admin := User{
		Name:         "System Admin",
		Email:        email,
		PasswordHash: string(hash),
		Role:         "admin",
		IsActive:     true,
	}
	if err := DB.Create(&admin).Error; err != nil {
		log.Printf("Warning: failed to seed admin user: %v", err)
		return
	}
	log.Printf("Seeded default admin user: %s", email)
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
