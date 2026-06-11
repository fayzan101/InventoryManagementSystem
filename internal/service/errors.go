package service

import (
	"fmt"
	"net/http"

	"myapp/pkg/httputil"
)

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string {
	return e.Message
}

func WriteHTTPError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*AppError); ok {
		httputil.Error(w, appErr.Status, appErr.Code, appErr.Message)
		return
	}
	httputil.InternalError(w, "An unexpected error occurred")
}

func NewError(status int, code, message string) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

func NotFound(entity string) *AppError {
	return NewError(404, "not_found", fmt.Sprintf("%s not found", entity))
}

func BadRequest(message string) *AppError {
	return NewError(400, "bad_request", message)
}

func Internal(message string) *AppError {
	return NewError(500, "internal_error", message)
}
