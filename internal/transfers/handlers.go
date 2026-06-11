package transfers

import (
	"encoding/json"
	"myapp/internal"
	"myapp/internal/auth"
	"myapp/internal/service"
	"myapp/pkg/httputil"
	"net/http"
	"strconv"
	"strings"
)

func transferSvc() *service.TransferService {
	return service.NewTransferService(internal.DB)
}

func CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromWarehouseID uint   `json:"from_warehouse_id"`
		ToWarehouseID   uint   `json:"to_warehouse_id"`
		Notes           string `json:"notes"`
		Items           []struct {
			ProductID uint `json:"product_id"`
			Quantity  int  `json:"quantity"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "Invalid request payload")
		return
	}

	items := make([]service.TransferLineInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.TransferLineInput{ProductID: item.ProductID, Quantity: item.Quantity}
	}

	transfer, err := transferSvc().CreateTransfer(service.CreateTransferInput{
		FromWarehouseID: req.FromWarehouseID,
		ToWarehouseID:   req.ToWarehouseID,
		Notes:           req.Notes,
		Items:           items,
		Actor:           auth.UserIDString(r.Context()),
	})
	if err != nil {
		service.WriteHTTPError(w, err)
		return
	}
	httputil.Success(w, http.StatusCreated, transfer)
}

func ListTransfers(w http.ResponseWriter, r *http.Request) {
	var transfers []internal.StockTransfer
	query := internal.DB.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Items.Product").
		Order("created_at DESC")
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if warehouseID := r.URL.Query().Get("warehouse_id"); warehouseID != "" {
		query = query.Where("from_warehouse_id = ? OR to_warehouse_id = ?", warehouseID, warehouseID)
	}

	paginated, meta, err := httputil.PaginateQuery(query, r)
	if err != nil {
		httputil.InternalError(w, "Failed to fetch transfers")
		return
	}
	if err := paginated.Find(&transfers).Error; err != nil {
		httputil.InternalError(w, "Failed to fetch transfers")
		return
	}
	httputil.Paginated(w, transfers, meta)
}

func GetTransfer(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/transfers/")
	if id == 0 {
		httputil.BadRequest(w, "Invalid transfer ID")
		return
	}

	var transfer internal.StockTransfer
	if err := internal.DB.Preload("FromWarehouse").Preload("ToWarehouse").
		Preload("Items.Product").First(&transfer, id).Error; err != nil {
		httputil.NotFound(w, "Transfer not found")
		return
	}
	httputil.Success(w, http.StatusOK, transfer)
}

func extractID(path, prefix string) int {
	idStr := strings.TrimPrefix(path, prefix)
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}
	id, _ := strconv.Atoi(idStr)
	return id
}
