package transfers

import (
	"encoding/json"
	"fmt"
	"myapp/internal"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromWarehouseID uint `json:"from_warehouse_id"`
		ToWarehouseID   uint `json:"to_warehouse_id"`
		Notes           string `json:"notes"`
		Items           []struct {
			ProductID uint `json:"product_id"`
			Quantity  int  `json:"quantity"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.FromWarehouseID == 0 || req.ToWarehouseID == 0 {
		http.Error(w, "Source and destination warehouses are required", http.StatusBadRequest)
		return
	}
	if req.FromWarehouseID == req.ToWarehouseID {
		http.Error(w, "Source and destination must differ", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "At least one item is required", http.StatusBadRequest)
		return
	}

	tx := internal.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range req.Items {
		var inv internal.Inventory
		if err := tx.Where("product_id = ? AND warehouse_id = ?", item.ProductID, req.FromWarehouseID).
			First(&inv).Error; err != nil || inv.Quantity < item.Quantity {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Insufficient stock for product %d in source warehouse", item.ProductID), http.StatusBadRequest)
			return
		}
	}

	now := time.Now()
	transfer := internal.StockTransfer{
		TransferNumber:  fmt.Sprintf("TRF-%d", now.Unix()),
		FromWarehouseID: req.FromWarehouseID,
		ToWarehouseID:   req.ToWarehouseID,
		Status:          "completed",
		Notes:           req.Notes,
		CreatedBy:       "system",
		CompletedAt:     &now,
	}
	if err := tx.Create(&transfer).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create transfer", http.StatusInternalServerError)
		return
	}

	for _, item := range req.Items {
		ti := internal.StockTransferItem{
			TransferID: transfer.ID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
		}
		if err := tx.Create(&ti).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to create transfer item", http.StatusInternalServerError)
			return
		}

		var src internal.Inventory
		tx.Where("product_id = ? AND warehouse_id = ?", item.ProductID, req.FromWarehouseID).First(&src)
		src.Quantity -= item.Quantity
		tx.Save(&src)

		var dst internal.Inventory
		err := tx.Where("product_id = ? AND warehouse_id = ?", item.ProductID, req.ToWarehouseID).First(&dst).Error
		if err != nil {
			dst = internal.Inventory{
				ProductID:   item.ProductID,
				WarehouseID: req.ToWarehouseID,
				Quantity:    item.Quantity,
			}
			tx.Create(&dst)
		} else {
			dst.Quantity += item.Quantity
			tx.Save(&dst)
		}

		ref := transfer.TransferNumber
		tx.Create(&internal.StockMovement{
			ProductID: item.ProductID, WarehouseID: req.FromWarehouseID,
			Type: "OUT", Quantity: -item.Quantity, Reference: ref,
			Reason: "Stock transfer out", CreatedBy: "system", CreatedAt: now,
		})
		tx.Create(&internal.StockMovement{
			ProductID: item.ProductID, WarehouseID: req.ToWarehouseID,
			Type: "IN", Quantity: item.Quantity, Reference: ref,
			Reason: "Stock transfer in", CreatedBy: "system", CreatedAt: now,
		})
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to complete transfer", http.StatusInternalServerError)
		return
	}

	internal.LogAudit("CREATE", "StockTransfer", transfer.ID, "system", "Completed stock transfer")

	internal.DB.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Items.Product").First(&transfer, transfer.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": transfer})
}

func ListTransfers(w http.ResponseWriter, r *http.Request) {
	var transfers []internal.StockTransfer
	query := internal.DB.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Items.Product").
		Order("created_at DESC").Limit(100)

	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if warehouseID := r.URL.Query().Get("warehouse_id"); warehouseID != "" {
		query = query.Where("from_warehouse_id = ? OR to_warehouse_id = ?", warehouseID, warehouseID)
	}

	if err := query.Find(&transfers).Error; err != nil {
		http.Error(w, "Failed to fetch transfers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": transfers})
}

func GetTransfer(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/transfers/")
	if id == 0 {
		http.Error(w, "Invalid transfer ID", http.StatusBadRequest)
		return
	}

	var transfer internal.StockTransfer
	if err := internal.DB.Preload("FromWarehouse").Preload("ToWarehouse").
		Preload("Items.Product").First(&transfer, id).Error; err != nil {
		http.Error(w, "Transfer not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": transfer})
}

func extractID(path, prefix string) int {
	idStr := strings.TrimPrefix(path, prefix)
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}
	id, _ := strconv.Atoi(idStr)
	return id
}
