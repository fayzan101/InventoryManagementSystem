package auth

import (
	"net/http"
	"strings"
)

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleStaff   = "staff"
)

func IsValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleManager, RoleStaff:
		return true
	default:
		return false
	}
}

// Allowed returns whether the role may access the HTTP method + path.
func Allowed(role, method, path string) bool {
	if role == RoleAdmin {
		return true
	}

	method = strings.ToUpper(method)
	path = strings.TrimSuffix(path, "/")

	if method == http.MethodGet || method == http.MethodOptions {
		return true
	}

	if role == RoleStaff {
		return method == http.MethodPost && path == "/orders"
	}

	// manager: write access except user management and deletes
	if role == RoleManager {
		if method == http.MethodDelete {
			return false
		}
		if strings.HasPrefix(path, "/auth/register") || strings.HasPrefix(path, "/employees") {
			return method == http.MethodGet
		}
		return true
	}

	return false
}
