package products

import (
	"encoding/json"
	"myapp/internal"
	"myapp/internal/auth"
	"myapp/internal/websocket"
	"myapp/pkg/httputil"
	"net/http"
	"strconv"
	"strings"
)

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product internal.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		httputil.BadRequest(w, "Invalid request payload")
		return
	}
	if err := httputil.Required("name", product.Name); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}
	if err := httputil.Required("sku", product.SKU); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}
	if err := httputil.NonNegativeFloat("price", product.Price); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

	if err := internal.DB.Create(&product).Error; err != nil {
		httputil.InternalError(w, "Failed to create product")
		return
	}
	internal.LogAudit("CREATE", "Product", product.ID, auth.UserIDString(r.Context()), "Created new product")

	// Broadcast product creation via WebSocket
	if hub := websocket.GetHub(); hub != nil {
		hub.BroadcastProductUpdate(product.ID, product.Name, product.SKU, product.Category, product.Price, "created")
	}

	httputil.Success(w, http.StatusCreated, product)
}

func ListProducts(w http.ResponseWriter, r *http.Request) {
	var products []internal.Product
	query := internal.DB.Model(&internal.Product{})
	if category := r.URL.Query().Get("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	paginated, meta, err := httputil.PaginateQuery(query.Order("name ASC"), r)
	if err != nil {
		httputil.InternalError(w, "Failed to fetch products")
		return
	}
	if err := paginated.Find(&products).Error; err != nil {
		httputil.InternalError(w, "Failed to fetch products")
		return
	}
	httputil.Paginated(w, products, meta)
}
func GetProduct(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/products/")
	if id == 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var product internal.Product
	if err := internal.DB.First(&product, id).Error; err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   product,
	})
}
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/products/")
	if id == 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var product internal.Product
	if err := internal.DB.First(&product, id).Error; err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	var updates internal.Product
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Store old price for price change alerts
	oldPrice := product.Price

	if err := internal.DB.Model(&product).Updates(updates).Error; err != nil {
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	internal.LogAudit("UPDATE", "Product", product.ID, "system", "Updated product")

	// Broadcast product update via WebSocket
	if hub := websocket.GetHub(); hub != nil {
		hub.BroadcastProductUpdate(product.ID, product.Name, product.SKU, product.Category, product.Price, "updated")
		
		// Send price alert if price changed by more than 10%
		if oldPrice > 0 && product.Price > 0 {
			changePercent := ((product.Price - oldPrice) / oldPrice) * 100
			if changePercent > 10 || changePercent < -10 {
				hub.BroadcastProductPriceAlert(product.ID, product.Name, oldPrice, product.Price, changePercent)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   product,
	})
}
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/products/")
	if id == 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Fetch product details before deletion for WebSocket broadcast
	var product internal.Product
	if err := internal.DB.First(&product, id).Error; err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	productName := product.Name
	productSKU := product.SKU
	productCategory := product.Category
	productPrice := product.Price
	productID := product.ID

	if err := internal.DB.Delete(&product).Error; err != nil {
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	internal.LogAudit("DELETE", "Product", uint(id), "system", "Deleted product")

	// Broadcast product deletion via WebSocket
	if hub := websocket.GetHub(); hub != nil {
		hub.BroadcastProductUpdate(productID, productName, productSKU, productCategory, productPrice, "deleted")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Product deleted successfully",
	})
}
func SearchProducts(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("q")
	if keyword == "" {
		http.Error(w, "Search keyword required", http.StatusBadRequest)
		return
	}

	var products []internal.Product
	searchPattern := "%" + keyword + "%"
	if err := internal.DB.Where("name LIKE ? OR sku LIKE ? OR description LIKE ?",
		searchPattern, searchPattern, searchPattern).Find(&products).Error; err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   products,
	})
}
func extractID(path, prefix string) int {
	idStr := strings.TrimPrefix(path, prefix)
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}
	id, _ := strconv.Atoi(idStr)
	return id
}
