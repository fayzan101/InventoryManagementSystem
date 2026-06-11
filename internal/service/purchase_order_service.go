package service

import (
	"fmt"
	"time"

	"myapp/internal"

	"gorm.io/gorm"
)

type POLineInput struct {
	ProductID uint
	Quantity  int
	UnitPrice float64
}

type CreatePOInput struct {
	SupplierID uint
	Items      []POLineInput
	Actor      string
}

type ReceivePOInput struct {
	POID        uint
	WarehouseID uint
	Actor       string
}

type PurchaseOrderService struct {
	DB *gorm.DB
}

func NewPurchaseOrderService(db *gorm.DB) *PurchaseOrderService {
	return &PurchaseOrderService{DB: db}
}

func (s *PurchaseOrderService) CreatePurchaseOrder(input CreatePOInput) (*internal.PurchaseOrder, error) {
	if len(input.Items) == 0 {
		return nil, BadRequest("at least one item is required")
	}

	poNumber := fmt.Sprintf("PO-%d", time.Now().UnixNano())
	var total float64
	for _, item := range input.Items {
		total += item.UnitPrice * float64(item.Quantity)
	}

	po := internal.PurchaseOrder{
		PONumber:   poNumber,
		SupplierID: input.SupplierID,
		Status:     "pending",
		TotalCost:  total,
		OrderDate:  time.Now(),
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&po).Error; err != nil {
			return Internal("failed to create purchase order")
		}
		for _, item := range input.Items {
			if err := tx.Create(&internal.POItem{
				POID: po.ID, ProductID: item.ProductID,
				Quantity: item.Quantity, UnitPrice: item.UnitPrice,
			}).Error; err != nil {
				return Internal("failed to create PO item")
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

	internal.LogAudit("CREATE", "PurchaseOrder", po.ID, input.Actor, "Created purchase order")
	return &po, nil
}

func (s *PurchaseOrderService) ReceivePurchaseOrder(input ReceivePOInput) (*internal.PurchaseOrder, error) {
	var po internal.PurchaseOrder
	if err := s.DB.Preload("Items").First(&po, input.POID).Error; err != nil {
		return nil, NotFound("Purchase order")
	}
	if po.Status == "received" {
		return nil, BadRequest("purchase order already received")
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for _, item := range po.Items {
			var inv internal.Inventory
			err := tx.Where("product_id = ? AND warehouse_id = ?",
				item.ProductID, input.WarehouseID).First(&inv).Error
			if err != nil {
				inv = internal.Inventory{
					ProductID: item.ProductID, WarehouseID: input.WarehouseID,
					Quantity: item.Quantity,
				}
				if err := tx.Create(&inv).Error; err != nil {
					return Internal("failed to create inventory record")
				}
			} else {
				inv.Quantity += item.Quantity
				if err := tx.Save(&inv).Error; err != nil {
					return Internal("failed to update inventory")
				}
			}

			if err := tx.Create(&internal.StockMovement{
				ProductID: item.ProductID, WarehouseID: input.WarehouseID,
				Type: "IN", Quantity: item.Quantity, Reference: po.PONumber,
				Reason: "Purchase order received", CreatedBy: input.Actor, CreatedAt: now,
			}).Error; err != nil {
				return Internal("failed to log stock movement")
			}
		}

		po.Status = "received"
		po.ReceivedAt = &now
		return tx.Save(&po).Error
	})
	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return nil, appErr
		}
		return nil, Internal(err.Error())
	}

	internal.LogAudit("RECEIVE", "PurchaseOrder", po.ID, input.Actor, "Purchase order received")
	return &po, nil
}
