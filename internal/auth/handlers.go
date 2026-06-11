package auth

import (
	"encoding/json"
	"net/http"

	"myapp/internal"
	"myapp/pkg/httputil"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.MethodNotAllowed(w)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "Invalid request payload")
		return
	}
	if err := httputil.Required("email", req.Email); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}
	if err := httputil.Required("password", req.Password); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

	var user internal.User
	if err := internal.DB.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		httputil.Unauthorized(w, "Invalid email or password")
		return
	}
	if !CheckPassword(user.PasswordHash, req.Password) {
		httputil.Unauthorized(w, "Invalid email or password")
		return
	}

	token, err := GenerateToken(user)
	if err != nil {
		httputil.InternalError(w, "Failed to generate token")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
		},
	})
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.MethodNotAllowed(w)
		return
	}

	if !AuthDisabled() {
		user, ok := UserFromContext(r.Context())
		if !ok || user.Role != RoleAdmin {
			httputil.Forbidden(w, "Only admins can register users")
			return
		}
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "Invalid request payload")
		return
	}
	if err := httputil.Required("name", req.Name); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}
	if err := httputil.Required("email", req.Email); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}
	if err := httputil.MinLen("password", req.Password, 6); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}
	if req.Role == "" {
		req.Role = RoleStaff
	}
	if !IsValidRole(req.Role) {
		httputil.BadRequest(w, "role must be admin, manager, or staff")
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		httputil.InternalError(w, "Failed to hash password")
		return
	}

	user := internal.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
		IsActive:     true,
	}
	if err := internal.DB.Create(&user).Error; err != nil {
		httputil.BadRequest(w, "Email already registered")
		return
	}

	internal.LogAudit("CREATE", "User", user.ID, authUserIDString(r), "Registered new user")

	httputil.Success(w, http.StatusCreated, map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

func Me(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		httputil.Unauthorized(w, "Not authenticated")
		return
	}

	var dbUser internal.User
	if err := internal.DB.First(&dbUser, user.ID).Error; err != nil {
		httputil.NotFound(w, "User not found")
		return
	}

	httputil.Success(w, http.StatusOK, map[string]interface{}{
		"id":    dbUser.ID,
		"name":  dbUser.Name,
		"email": dbUser.Email,
		"role":  dbUser.Role,
	})
}

func authUserIDString(r *http.Request) string {
	return UserIDString(r.Context())
}
