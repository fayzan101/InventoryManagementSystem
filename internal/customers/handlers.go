package customers

import (
	"encoding/json"
	"myapp/internal"
	"net/http"
	"strconv"
	"strings"
)

func CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer internal.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if customer.Name == "" {
		http.Error(w, "Customer name is required", http.StatusBadRequest)
		return
	}

	if err := internal.DB.Create(&customer).Error; err != nil {
		http.Error(w, "Failed to create customer", http.StatusInternalServerError)
		return
	}
	internal.LogAudit("CREATE", "Customer", customer.ID, "system", "Created new customer")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": customer})
}

func ListCustomers(w http.ResponseWriter, r *http.Request) {
	var customers []internal.Customer
	query := internal.DB
	if company := r.URL.Query().Get("company"); company != "" {
		query = query.Where("company LIKE ?", "%"+company+"%")
	}
	if search := r.URL.Query().Get("q"); search != "" {
		pattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR email LIKE ? OR company LIKE ?", pattern, pattern, pattern)
	}

	if err := query.Order("name ASC").Find(&customers).Error; err != nil {
		http.Error(w, "Failed to fetch customers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": customers})
}

func GetCustomer(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/customers/")
	if id == 0 {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	var customer internal.Customer
	if err := internal.DB.First(&customer, id).Error; err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": customer})
}

func UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/customers/")
	if id == 0 {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	var customer internal.Customer
	if err := internal.DB.First(&customer, id).Error; err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	var updates internal.Customer
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := internal.DB.Model(&customer).Updates(updates).Error; err != nil {
		http.Error(w, "Failed to update customer", http.StatusInternalServerError)
		return
	}
	internal.LogAudit("UPDATE", "Customer", customer.ID, "system", "Updated customer")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": customer})
}

func DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/customers/")
	if id == 0 {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	var customer internal.Customer
	if err := internal.DB.First(&customer, id).Error; err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	if err := internal.DB.Delete(&customer).Error; err != nil {
		http.Error(w, "Failed to delete customer", http.StatusInternalServerError)
		return
	}
	internal.LogAudit("DELETE", "Customer", uint(id), "system", "Deleted customer")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Customer deleted successfully",
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
