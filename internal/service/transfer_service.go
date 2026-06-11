package service

import (
	"fmt"
	"time"

	"myapp/internal"

	"gorm.io/gorm"
)

type TransferLineInput struct {
	ProductID uint
	Quantity  int
}

type CreateTransferInput struct {
	FromWarehouseID uint
	ToWarehouseID   uint
	Notes           string
	Items           []TransferLineInput
	Actor           string
}

type TransferService struct {
	DB *gorm.DB
}

func NewTransferService(db *gorm.DB) *TransferService {
	return &TransferService{DB: db}
}

func (s *TransferService) CreateTransfer(input CreateTransferInput) (*internal.StockTransfer, error) {
	if input.FromWarehouseID == input.ToWarehouseID {
		return nil, BadRequest("source and destination must differ")
	}
	if len(input.Items) == 0 {
		return nil, BadRequest("at least one item is required")
	}

	now := time.Now()
	transfer := internal.StockTransfer{
		TransferNumber:  fmt.Sprintf("TRF-%d", now.UnixNano()),
		FromWarehouseID: input.FromWarehouseID,
		ToWarehouseID:   input.ToWarehouseID,
		Status:          "completed",
		Notes:           input.Notes,
		CreatedBy:       input.Actor,
		CompletedAt:     &now,
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range input.Items {
			var inv internal.Inventory
			if err := tx.Where("product_id = ? AND warehouse_id = ?",
				item.ProductID, input.FromWarehouseID).First(&inv).Error; err != nil {
				return BadRequest(fmt.Sprintf("no stock for product %d in source warehouse", item.ProductID))
			}
			if inv.Quantity < item.Quantity {
				return BadRequest(fmt.Sprintf("insufficient stock for product %d", item.ProductID))
			}
		}

		if err := tx.Create(&transfer).Error; err != nil {
			return Internal("failed to create transfer")
		}

		for _, item := range input.Items {
			if err := tx.Create(&internal.StockTransferItem{
				TransferID: transfer.ID, ProductID: item.ProductID, Quantity: item.Quantity,
			}).Error; err != nil {
				return Internal("failed to create transfer item")
			}

			result := tx.Model(&internal.Inventory{}).
				Where("product_id = ? AND warehouse_id = ? AND quantity >= ?",
					item.ProductID, input.FromWarehouseID, item.Quantity).
				Update("quantity", gorm.Expr("quantity - ?", item.Quantity))
			if result.Error != nil || result.RowsAffected == 0 {
				return BadRequest("insufficient stock during transfer")
			}

			var dst internal.Inventory
			err := tx.Where("product_id = ? AND warehouse_id = ?",
				item.ProductID, input.ToWarehouseID).First(&dst).Error
			if err != nil {
				dst = internal.Inventory{
					ProductID: item.ProductID, WarehouseID: input.ToWarehouseID,
					Quantity: item.Quantity,
				}
				if err := tx.Create(&dst).Error; err != nil {
					return Internal("failed to create destination inventory")
				}
			} else {
				dst.Quantity += item.Quantity
				if err := tx.Save(&dst).Error; err != nil {
					return Internal("failed to update destination inventory")
				}
			}

			ref := transfer.TransferNumber
			if err := tx.Create(&internal.StockMovement{
				ProductID: item.ProductID, WarehouseID: input.FromWarehouseID,
				Type: "OUT", Quantity: -item.Quantity, Reference: ref,
				Reason: "Stock transfer out", CreatedBy: input.Actor, CreatedAt: now,
			}).Error; err != nil {
				return Internal("failed to log outbound movement")
			}
			if err := tx.Create(&internal.StockMovement{
				ProductID: item.ProductID, WarehouseID: input.ToWarehouseID,
				Type: "IN", Quantity: item.Quantity, Reference: ref,
				Reason: "Stock transfer in", CreatedBy: input.Actor, CreatedAt: now,
			}).Error; err != nil {
				return Internal("failed to log inbound movement")
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

	internal.LogAudit("CREATE", "StockTransfer", transfer.ID, input.Actor, "Completed stock transfer")
	s.DB.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Items.Product").First(&transfer, transfer.ID)
	return &transfer, nil
}
