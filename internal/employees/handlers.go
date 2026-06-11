package employees

import (
	"encoding/json"
	"myapp/internal"
	"net/http"
	"strconv"
	"strings"
)

func CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var employee internal.Employee
	if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if employee.Name == "" || employee.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}
	if employee.Role == "" {
		employee.Role = "staff"
	}

	if err := internal.DB.Create(&employee).Error; err != nil {
		http.Error(w, "Failed to create employee", http.StatusInternalServerError)
		return
	}
	internal.LogAudit("CREATE", "Employee", employee.ID, "system", "Created new employee")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": employee})
}

func ListEmployees(w http.ResponseWriter, r *http.Request) {
	var employees []internal.Employee
	query := internal.DB.Preload("Warehouse")
	if role := r.URL.Query().Get("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	if warehouseID := r.URL.Query().Get("warehouse_id"); warehouseID != "" {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	if active := r.URL.Query().Get("active"); active == "true" {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("name ASC").Find(&employees).Error; err != nil {
		http.Error(w, "Failed to fetch employees", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": employees})
}

func GetEmployee(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/employees/")
	if id == 0 {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	var employee internal.Employee
	if err := internal.DB.Preload("Warehouse").First(&employee, id).Error; err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": employee})
}

func UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/employees/")
	if id == 0 {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	var employee internal.Employee
	if err := internal.DB.First(&employee, id).Error; err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	var updates internal.Employee
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := internal.DB.Model(&employee).Updates(updates).Error; err != nil {
		http.Error(w, "Failed to update employee", http.StatusInternalServerError)
		return
	}
	internal.LogAudit("UPDATE", "Employee", employee.ID, "system", "Updated employee")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": employee})
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/employees/")
	if id == 0 {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	if err := internal.DB.Delete(&internal.Employee{}, id).Error; err != nil {
		http.Error(w, "Failed to delete employee", http.StatusInternalServerError)
		return
	}
	internal.LogAudit("DELETE", "Employee", uint(id), "system", "Deleted employee")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Employee deleted successfully",
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
