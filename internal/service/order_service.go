package service

import (
	"fmt"
	"time"

	"myapp/internal"
	"myapp/pkg/httputil"

	"gorm.io/gorm"
)

type OrderLineInput struct {
	ProductID   uint
	WarehouseID uint
	Quantity    int
}

type CreateOrderInput struct {
	CustomerID    *uint
	CustomerName  string
	CustomerEmail string
	Items         []OrderLineInput
	Actor         string
}

type OrderService struct {
	DB *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{DB: db}
}

func (s *OrderService) CreateOrder(input CreateOrderInput) (*internal.Order, error) {
	if len(input.Items) == 0 {
		return nil, BadRequest("at least one order item is required")
	}
	for _, item := range input.Items {
		if err := httputil.PositiveInt("quantity", item.Quantity); err != nil {
			return nil, BadRequest(err.Error())
		}
	}

	orderNumber := fmt.Sprintf("ORD-%d", time.Now().UnixNano())
	order := internal.Order{
		OrderNumber:   orderNumber,
		CustomerID:    input.CustomerID,
		CustomerName:  input.CustomerName,
		CustomerEmail: input.CustomerEmail,
		Status:        "pending",
		OrderDate:     time.Now(),
	}

	if input.CustomerID != nil {
		var customer internal.Customer
		if err := s.DB.First(&customer, *input.CustomerID).Error; err != nil {
			return nil, NotFound("Customer")
		}
		order.CustomerName = customer.Name
		order.CustomerEmail = customer.Email
	} else if input.CustomerName == "" {
		return nil, BadRequest("customer_name or customer_id is required")
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var total float64
		for _, item := range input.Items {
			var product internal.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return NotFound("Product")
			}

			var inv internal.Inventory
			if err := tx.Where("product_id = ? AND warehouse_id = ?", item.ProductID, item.WarehouseID).
				First(&inv).Error; err != nil {
				return BadRequest("product not available in warehouse")
			}
			if inv.Quantity < item.Quantity {
				return BadRequest("insufficient stock")
			}
			total += product.Price * float64(item.Quantity)
		}

		order.TotalAmount = total
		if err := tx.Create(&order).Error; err != nil {
			return Internal("failed to create order")
		}

		now := time.Now()
		for _, item := range input.Items {
			var product internal.Product
			_ = tx.First(&product, item.ProductID)

			if err := tx.Create(&internal.OrderItem{
				OrderID: order.ID, ProductID: item.ProductID,
				WarehouseID: item.WarehouseID, Quantity: item.Quantity,
				UnitPrice: product.Price,
			}).Error; err != nil {
				return Internal("failed to create order item")
			}

			result := tx.Model(&internal.Inventory{}).
				Where("product_id = ? AND warehouse_id = ? AND quantity >= ?",
					item.ProductID, item.WarehouseID, item.Quantity).
				Update("quantity", gorm.Expr("quantity - ?", item.Quantity))
			if result.Error != nil {
				return Internal("failed to update inventory")
			}
			if result.RowsAffected == 0 {
				return BadRequest("insufficient stock")
			}

			if err := tx.Create(&internal.StockMovement{
				ProductID: item.ProductID, WarehouseID: item.WarehouseID,
				Type: "OUT", Quantity: -item.Quantity, Reference: orderNumber,
				Reason: "Sales order", CreatedBy: input.Actor, CreatedAt: now,
			}).Error; err != nil {
				return Internal("failed to log stock movement")
			}
		}
		return nil
	})
	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return nil, appErr
		}
		return nil, Internal(err.Error())
	}

	internal.LogAudit("CREATE", "Order", order.ID, input.Actor, "Created new order")
	return &order, nil
}

func (s *OrderService) GetInventoryQty(productID, warehouseID uint) (int, error) {
	var inv internal.Inventory
	if err := s.DB.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		First(&inv).Error; err != nil {
		return 0, err
	}
	return inv.Quantity, nil
}
