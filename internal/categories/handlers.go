package categories

import (
	"encoding/json"
	"myapp/internal"
	"net/http"
	"strconv"
	"strings"
)

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category internal.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if category.Name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}

	if err := internal.DB.Create(&category).Error; err != nil {
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}
	internal.LogAudit("CREATE", "Category", category.ID, "system", "Created new category")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": category})
}

func ListCategories(w http.ResponseWriter, r *http.Request) {
	var categories []internal.Category
	query := internal.DB
	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		query = query.Where("parent_id = ?", parentID)
	}

	if err := query.Order("name ASC").Find(&categories).Error; err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": categories})
}

func GetCategory(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/categories/")
	if id == 0 {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var category internal.Category
	if err := internal.DB.First(&category, id).Error; err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": category})
}

func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/categories/")
	if id == 0 {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var category internal.Category
	if err := internal.DB.First(&category, id).Error; err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	var updates internal.Category
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := internal.DB.Model(&category).Updates(updates).Error; err != nil {
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}
	internal.LogAudit("UPDATE", "Category", category.ID, "system", "Updated category")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": category})
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/categories/")
	if id == 0 {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var count int64
	internal.DB.Model(&internal.Product{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		http.Error(w, "Cannot delete category with linked products", http.StatusBadRequest)
		return
	}

	if err := internal.DB.Delete(&internal.Category{}, id).Error; err != nil {
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}
	internal.LogAudit("DELETE", "Category", uint(id), "system", "Deleted category")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Category deleted successfully",
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
