package reports

import (
	"encoding/json"
	"myapp/internal"
	"net/http"
)

func GetStockSummary(w http.ResponseWriter, r *http.Request) {
	var results []struct {
		WarehouseName string  `json:"warehouse_name"`
		ProductName   string  `json:"product_name"`
		SKU           string  `json:"sku"`
		Quantity      int     `json:"quantity"`
		Value         float64 `json:"value"`
	}

	query := `
		SELECT 
			w.name as warehouse_name,
			p.name as product_name,
			p.sku,
			i.quantity,
			(i.quantity * p.price) as value
		FROM inventories i
		JOIN products p ON i.product_id = p.id
		JOIN warehouses w ON i.warehouse_id = w.id
		ORDER BY w.name, p.name
	`

	if err := internal.DB.Raw(query).Scan(&results).Error; err != nil {
		http.Error(w, "Failed to generate stock summary", http.StatusInternalServerError)
		return
	}
	var totalValue float64
	var totalItems int
	for _, r := range results {
		totalValue += r.Value
		totalItems += r.Quantity
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"items":       results,
			"total_value": totalValue,
			"total_items": totalItems,
		},
	})
}
func GetSalesSummary(w http.ResponseWriter, r *http.Request) {
	var byStatus []struct {
		Status string  `json:"status"`
		Count  int     `json:"count"`
		Total  float64 `json:"total"`
	}

	statusQuery := `
		SELECT status, COUNT(*) as count, COALESCE(SUM(total_amount), 0) as total
		FROM orders
		GROUP BY status
		ORDER BY status
	`
	if err := internal.DB.Raw(statusQuery).Scan(&byStatus).Error; err != nil {
		http.Error(w, "Failed to generate sales summary", http.StatusInternalServerError)
		return
	}

	var overall struct {
		TotalOrders   int     `json:"total_orders"`
		TotalRevenue  float64 `json:"total_revenue"`
		AverageOrder  float64 `json:"average_order"`
	}
	internal.DB.Raw(`SELECT COUNT(*) as total_orders, COALESCE(SUM(total_amount), 0) as total_revenue FROM orders`).Scan(&overall)
	if overall.TotalOrders > 0 {
		overall.AverageOrder = overall.TotalRevenue / float64(overall.TotalOrders)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"by_status": byStatus,
			"overall":   overall,
		},
	})
}

func GetTopProducts(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "10"
	}

	var results []struct {
		ProductID   uint    `json:"product_id"`
		ProductName string  `json:"product_name"`
		SKU         string  `json:"sku"`
		UnitsSold   int     `json:"units_sold"`
		Revenue     float64 `json:"revenue"`
	}

	query := `
		SELECT
			p.id as product_id,
			p.name as product_name,
			p.sku,
			SUM(oi.quantity) as units_sold,
			SUM(oi.quantity * oi.unit_price) as revenue
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		GROUP BY p.id, p.name, p.sku
		ORDER BY units_sold DESC
		LIMIT ?
	`
	if err := internal.DB.Raw(query, limit).Scan(&results).Error; err != nil {
		http.Error(w, "Failed to generate top products report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   results,
	})
}

func GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	var logs []internal.AuditLog
	query := internal.DB.Order("created_at DESC").Limit(100)
	entity := r.URL.Query().Get("entity")
	if entity != "" {
		query = query.Where("entity = ?", entity)
	}
	action := r.URL.Query().Get("action")
	if action != "" {
		query = query.Where("action = ?", action)
	}

	if err := query.Find(&logs).Error; err != nil {
		http.Error(w, "Failed to fetch audit logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   logs,
	})
}
