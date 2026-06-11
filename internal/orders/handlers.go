package orders

import (
	"encoding/json"
	"fmt"
	"myapp/internal"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func CreatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SupplierID uint `json:"supplier_id"`
		Items      []struct {
			ProductID uint    `json:"product_id"`
			Quantity  int     `json:"quantity"`
			UnitPrice float64 `json:"unit_price"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	poNumber := fmt.Sprintf("PO-%d", time.Now().Unix())
	var total float64
	for _, item := range req.Items {
		total += item.UnitPrice * float64(item.Quantity)
	}

	po := internal.PurchaseOrder{
		PONumber:   poNumber,
		SupplierID: req.SupplierID,
		Status:     "pending",
		TotalCost:  total,
		OrderDate:  time.Now(),
	}

	if err := internal.DB.Create(&po).Error; err != nil {
		http.Error(w, "Failed to create purchase order", http.StatusInternalServerError)
		return
	}
	for _, item := range req.Items {
		poItem := internal.POItem{
			POID:      po.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
		internal.DB.Create(&poItem)
	}

	internal.LogAudit("CREATE", "PurchaseOrder", po.ID, "system", "Created new purchase order")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   po,
	})
}
func ListPurchaseOrders(w http.ResponseWriter, r *http.Request) {
	var pos []internal.PurchaseOrder
	query := internal.DB.Preload("Supplier").Preload("Items.Product")
	status := r.URL.Query().Get("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&pos).Error; err != nil {
		http.Error(w, "Failed to fetch purchase orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   pos,
	})
}
func ReceivePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/purchase-orders/")
	if id == 0 {
		http.Error(w, "Invalid purchase order ID", http.StatusBadRequest)
		return
	}

	var req struct {
		WarehouseID uint `json:"warehouse_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var po internal.PurchaseOrder
	if err := internal.DB.Preload("Items").First(&po, id).Error; err != nil {
		http.Error(w, "Purchase order not found", http.StatusNotFound)
		return
	}
	for _, item := range po.Items {
		var inv internal.Inventory
		err := internal.DB.Where("product_id = ? AND warehouse_id = ?",
			item.ProductID, req.WarehouseID).First(&inv).Error

		if err != nil {
			inv = internal.Inventory{
				ProductID:   item.ProductID,
				WarehouseID: req.WarehouseID,
				Quantity:    item.Quantity,
			}
			internal.DB.Create(&inv)
		} else {
			inv.Quantity += item.Quantity
			internal.DB.Save(&inv)
		}
		movement := internal.StockMovement{
			ProductID:   item.ProductID,
			WarehouseID: req.WarehouseID,
			Type:        "IN",
			Quantity:    item.Quantity,
			Reference:   po.PONumber,
			Reason:      "Purchase order received",
			CreatedBy:   "system",
			CreatedAt:   time.Now(),
		}
		internal.DB.Create(&movement)
	}
	now := time.Now()
	po.Status = "received"
	po.ReceivedAt = &now
	internal.DB.Save(&po)

	internal.LogAudit("RECEIVE", "PurchaseOrder", po.ID, "system", "Purchase order received")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Purchase order received successfully",
		"data":    po,
	})
}
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID    *uint  `json:"customer_id"`
		CustomerName  string `json:"customer_name"`
		CustomerEmail string `json:"customer_email"`
		Items         []struct {
			ProductID   uint `json:"product_id"`
			WarehouseID uint `json:"warehouse_id"`
			Quantity    int  `json:"quantity"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	orderNumber := fmt.Sprintf("ORD-%d", time.Now().Unix())
	var total float64
	for _, item := range req.Items {
		var product internal.Product
		if err := internal.DB.First(&product, item.ProductID).Error; err != nil {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		var inv internal.Inventory
		if err := internal.DB.Where("product_id = ? AND warehouse_id = ?",
			item.ProductID, item.WarehouseID).First(&inv).Error; err != nil {
			http.Error(w, "Product not available in warehouse", http.StatusBadRequest)
			return
		}

		if inv.Quantity < item.Quantity {
			http.Error(w, "Insufficient stock", http.StatusBadRequest)
			return
		}

		total += product.Price * float64(item.Quantity)
	}

	order := internal.Order{
		OrderNumber:   orderNumber,
		CustomerID:    req.CustomerID,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		Status:        "pending",
		TotalAmount:   total,
		OrderDate:     time.Now(),
	}

	if req.CustomerID != nil {
		var customer internal.Customer
		if err := internal.DB.First(&customer, *req.CustomerID).Error; err != nil {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}
		order.CustomerName = customer.Name
		order.CustomerEmail = customer.Email
	} else if req.CustomerName == "" {
		http.Error(w, "customer_name or customer_id is required", http.StatusBadRequest)
		return
	}

	if err := internal.DB.Create(&order).Error; err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}
	for _, item := range req.Items {
		var product internal.Product
		internal.DB.First(&product, item.ProductID)

		orderItem := internal.OrderItem{
			OrderID:     order.ID,
			ProductID:   item.ProductID,
			WarehouseID: item.WarehouseID,
			Quantity:    item.Quantity,
			UnitPrice:   product.Price,
		}
		internal.DB.Create(&orderItem)
		var inv internal.Inventory
		internal.DB.Where("product_id = ? AND warehouse_id = ?",
			item.ProductID, item.WarehouseID).First(&inv)
		inv.Quantity -= item.Quantity
		internal.DB.Save(&inv)
		movement := internal.StockMovement{
			ProductID:   item.ProductID,
			WarehouseID: item.WarehouseID,
			Type:        "OUT",
			Quantity:    -item.Quantity,
			Reference:   orderNumber,
			Reason:      "Sales order",
			CreatedBy:   "system",
			CreatedAt:   time.Now(),
		}
		internal.DB.Create(&movement)
	}

	internal.LogAudit("CREATE", "Order", order.ID, "system", "Created new order")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   order,
	})
}
func ListOrders(w http.ResponseWriter, r *http.Request) {
	var orders []internal.Order
	query := internal.DB.Preload("Items.Product")
	status := r.URL.Query().Get("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   orders,
	})
}
func GetOrder(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/orders/")
	if id == 0 {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var order internal.Order
	if err := internal.DB.Preload("Items.Product").Preload("Customer").First(&order, id).Error; err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   order,
	})
}

func UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/orders/")
	if id == 0 {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var order internal.Order
	if err := internal.DB.First(&order, id).Error; err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	order.Status = req.Status
	now := time.Now()

	switch req.Status {
	case "shipped":
		order.ShippedAt = &now
	case "delivered":
		order.DeliveredAt = &now
	}

	internal.DB.Save(&order)

	internal.LogAudit("UPDATE_STATUS", "Order", order.ID, "system", "Updated order status to "+req.Status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   order,
	})
}

func extractID(path, prefix string) int {
	idStr := strings.TrimPrefix(path, prefix)
	parts := strings.Split(idStr, "/")
	id, _ := strconv.Atoi(parts[0])
	return id
}
