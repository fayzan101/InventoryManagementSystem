package orders

import (
	"encoding/json"
	"myapp/internal"
	"myapp/internal/auth"
	"myapp/internal/service"
	"myapp/pkg/httputil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func orderSvc() *service.OrderService {
	return service.NewOrderService(internal.DB)
}

func poSvc() *service.PurchaseOrderService {
	return service.NewPurchaseOrderService(internal.DB)
}

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
		httputil.BadRequest(w, "Invalid request payload")
		return
	}

	items := make([]service.POLineInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.POLineInput{
			ProductID: item.ProductID, Quantity: item.Quantity, UnitPrice: item.UnitPrice,
		}
	}

	po, err := poSvc().CreatePurchaseOrder(service.CreatePOInput{
		SupplierID: req.SupplierID,
		Items:      items,
		Actor:      auth.UserIDString(r.Context()),
	})
	if err != nil {
		service.WriteHTTPError(w, err)
		return
	}
	httputil.Success(w, http.StatusCreated, po)
}

func ListPurchaseOrders(w http.ResponseWriter, r *http.Request) {
	var pos []internal.PurchaseOrder
	query := internal.DB.Preload("Supplier").Preload("Items.Product")
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	paginated, meta, err := httputil.PaginateQuery(query.Order("created_at DESC"), r)
	if err != nil {
		httputil.InternalError(w, "Failed to fetch purchase orders")
		return
	}
	if err := paginated.Find(&pos).Error; err != nil {
		httputil.InternalError(w, "Failed to fetch purchase orders")
		return
	}
	httputil.Paginated(w, pos, meta)
}

func ReceivePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/purchase-orders/")
	if id == 0 {
		httputil.BadRequest(w, "Invalid purchase order ID")
		return
	}

	var req struct {
		WarehouseID uint `json:"warehouse_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "Invalid request payload")
		return
	}

	po, err := poSvc().ReceivePurchaseOrder(service.ReceivePOInput{
		POID: uint(id), WarehouseID: req.WarehouseID,
		Actor: auth.UserIDString(r.Context()),
	})
	if err != nil {
		service.WriteHTTPError(w, err)
		return
	}
	httputil.SuccessMessage(w, http.StatusOK, "Purchase order received successfully", po)
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
		httputil.BadRequest(w, "Invalid request payload")
		return
	}

	items := make([]service.OrderLineInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.OrderLineInput{
			ProductID: item.ProductID, WarehouseID: item.WarehouseID, Quantity: item.Quantity,
		}
	}

	order, err := orderSvc().CreateOrder(service.CreateOrderInput{
		CustomerID: req.CustomerID, CustomerName: req.CustomerName,
		CustomerEmail: req.CustomerEmail, Items: items,
		Actor: auth.UserIDString(r.Context()),
	})
	if err != nil {
		service.WriteHTTPError(w, err)
		return
	}
	httputil.Success(w, http.StatusCreated, order)
}

func ListOrders(w http.ResponseWriter, r *http.Request) {
	var orders []internal.Order
	query := internal.DB.Preload("Items.Product")
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	paginated, meta, err := httputil.PaginateQuery(query.Order("created_at DESC"), r)
	if err != nil {
		httputil.InternalError(w, "Failed to fetch orders")
		return
	}
	if err := paginated.Find(&orders).Error; err != nil {
		httputil.InternalError(w, "Failed to fetch orders")
		return
	}
	httputil.Paginated(w, orders, meta)
}

func GetOrder(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/orders/")
	if id == 0 {
		httputil.BadRequest(w, "Invalid order ID")
		return
	}

	var order internal.Order
	if err := internal.DB.Preload("Items.Product").Preload("Customer").First(&order, id).Error; err != nil {
		httputil.NotFound(w, "Order not found")
		return
	}
	httputil.Success(w, http.StatusOK, order)
}

func UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/orders/")
	if id == 0 {
		httputil.BadRequest(w, "Invalid order ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "Invalid request payload")
		return
	}
	if err := httputil.OneOf("status", req.Status, "pending", "processing", "shipped", "delivered", "cancelled"); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

	var order internal.Order
	if err := internal.DB.First(&order, id).Error; err != nil {
		httputil.NotFound(w, "Order not found")
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

	if err := internal.DB.Save(&order).Error; err != nil {
		httputil.InternalError(w, "Failed to update order status")
		return
	}
	internal.LogAudit("UPDATE_STATUS", "Order", order.ID, auth.UserIDString(r.Context()), "Updated order status to "+req.Status)
	httputil.Success(w, http.StatusOK, order)
}

func extractID(path, prefix string) int {
	idStr := strings.TrimPrefix(path, prefix)
	parts := strings.Split(idStr, "/")
	id, _ := strconv.Atoi(parts[0])
	return id
}
