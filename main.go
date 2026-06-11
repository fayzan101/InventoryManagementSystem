/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"myapp/internal"
	"myapp/internal/inventory"
	"myapp/internal/orders"
	"myapp/internal/products"
	"myapp/internal/reports"
	"myapp/internal/suppliers"
	"myapp/internal/swagger"
	"myapp/internal/warehouses"
	"myapp/internal/websocket"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

//go:embed docs/openapi.yaml
var openAPISpec []byte

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables.")
	}

	internal.InitDB(dbConnString())

	// Initialize WebSocket hub for real-time updates
	websocket.InitHub()
	log.Println("🔌 WebSocket hub initialized for real-time inventory updates")

	http.HandleFunc("/products", handleProducts)
	http.HandleFunc("/products/", handleProductsWithID)
	http.HandleFunc("/products/search", products.SearchProducts)
	http.HandleFunc("/warehouses", handleWarehouses)
	http.HandleFunc("/warehouses/", handleWarehousesWithID)
	http.HandleFunc("/inventory", handleInventory)
	http.HandleFunc("/inventory/", handleInventoryWithID)
	http.HandleFunc("/inventory/adjust", inventory.AdjustInventory)
	http.HandleFunc("/inventory/low-stock", inventory.GetLowStock)
	http.HandleFunc("/inventory/movements", inventory.GetStockMovements)

	// WebSocket endpoints for real-time updates
	http.HandleFunc("/ws/inventory", websocket.HandleWebSocket)
	http.HandleFunc("/ws/warehouses", websocket.HandleWebSocket)
	http.HandleFunc("/ws/products", websocket.HandleWebSocket)
	http.HandleFunc("/ws/suppliers", websocket.HandleWebSocket)

	http.HandleFunc("/suppliers", handleSuppliers)
	http.HandleFunc("/suppliers/", handleSuppliersWithID)
	http.HandleFunc("/purchase-orders", handlePurchaseOrders)
	http.HandleFunc("/purchase-orders/", handlePurchaseOrdersWithID)
	http.HandleFunc("/orders", handleOrders)
	http.HandleFunc("/orders/", handleOrdersWithID)
	http.HandleFunc("/reports/stock-summary", reports.GetStockSummary)
	http.HandleFunc("/audit-logs", reports.GetAuditLogs)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "IMS Backend API v1.0",
		})
	})

	swagger.Register(openAPISpec)

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = os.Getenv("APP_PORT")
	}
	if serverPort == "" {
		serverPort = "3000"
	}

	log.Printf("🚀 Inventory Management System API started on :%s", serverPort)
	log.Printf("📖 Swagger UI: http://localhost:%s/swagger/", serverPort)
	log.Println("📦 Total APIs: 25")
	log.Printf("🔌 WebSocket endpoints:")
	log.Printf("   - ws://localhost:%s/ws/inventory", serverPort)
	log.Printf("   - ws://localhost:%s/ws/warehouses", serverPort)
	log.Printf("   - ws://localhost:%s/ws/products", serverPort)
	log.Printf("   - ws://localhost:%s/ws/suppliers", serverPort)
	if err := http.ListenAndServe(":"+serverPort, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func dbConnString() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		sslmode,
	)
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		products.CreateProduct(w, r)
	case http.MethodGet:
		products.ListProducts(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProductsWithID(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/search") {
		products.SearchProducts(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		products.GetProduct(w, r)
	case http.MethodPut:
		products.UpdateProduct(w, r)
	case http.MethodDelete:
		products.DeleteProduct(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleWarehouses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		warehouses.CreateWarehouse(w, r)
	case http.MethodGet:
		warehouses.ListWarehouses(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleWarehousesWithID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		warehouses.GetWarehouse(w, r)
	case http.MethodPut:
		warehouses.UpdateWarehouse(w, r)
	case http.MethodDelete:
		warehouses.DeleteWarehouse(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		inventory.GetInventory(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleInventoryWithID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		inventory.GetProductInventory(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSuppliers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		suppliers.CreateSupplier(w, r)
	case http.MethodGet:
		suppliers.ListSuppliers(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSuppliersWithID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		suppliers.UpdateSupplier(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePurchaseOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		orders.CreatePurchaseOrder(w, r)
	case http.MethodGet:
		orders.ListPurchaseOrders(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePurchaseOrdersWithID(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "/receive") && r.Method == http.MethodPut {
		orders.ReceivePurchaseOrder(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		orders.CreateOrder(w, r)
	case http.MethodGet:
		orders.ListOrders(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleOrdersWithID(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "/status") && r.Method == http.MethodPut {
		orders.UpdateOrderStatus(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
