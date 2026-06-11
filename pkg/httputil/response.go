package httputil

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaginatedMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Success(w http.ResponseWriter, status int, data interface{}) {
	JSON(w, status, map[string]interface{}{"status": "success", "data": data})
}

func SuccessMessage(w http.ResponseWriter, status int, message string, data interface{}) {
	payload := map[string]interface{}{"status": "success", "message": message}
	if data != nil {
		payload["data"] = data
	}
	JSON(w, status, payload)
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorBody{Status: "error", Code: code, Message: message})
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "bad_request", message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "unauthorized", message)
}

func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "forbidden", message)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "not_found", message)
}

func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "internal_error", message)
}

func MethodNotAllowed(w http.ResponseWriter) {
	Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
}

func Paginated(w http.ResponseWriter, data interface{}, meta PaginatedMeta) {
	JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
		"meta":   meta,
	})
}
