package auth

import (
	"net/http"
	"os"
	"strings"

	"myapp/pkg/httputil"
)

func AuthDisabled() bool {
	return os.Getenv("AUTH_DISABLED") == "true"
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if AuthDisabled() {
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), AuthUser{
				ID: 0, Email: "system", Role: RoleAdmin,
			})))
			return
		}

		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			httputil.Unauthorized(w, "Missing or invalid Authorization header")
			return
		}

		claims, err := ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			httputil.Unauthorized(w, "Invalid or expired token")
			return
		}

		user := AuthUser{ID: claims.UserID, Email: claims.Email, Role: claims.Role}
		if !Allowed(user.Role, r.Method, r.URL.Path) {
			httputil.Forbidden(w, "Insufficient permissions for this action")
			return
		}

		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
	})
}

func PublicOnly(next http.Handler) http.Handler {
	return next
}
