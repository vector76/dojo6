package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"dojo6/backend/internal/models"
)

// Handler holds dependencies for auth HTTP handlers.
type Handler struct {
	Users  *models.UserRepository
	JWTSvc *JWTService
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type setupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

type authResponse struct {
	Token string   `json:"token"`
	User  userJSON `json:"user"`
}

type userJSON struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

func toUserJSON(u *models.User) userJSON {
	return userJSON{
		ID:    strconv.FormatInt(u.ID, 10),
		Email: u.Email,
		Name:  u.Name,
		Phone: u.Phone,
		Role:  u.Role,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// Login handles POST /api/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	var user models.User
	// GetByEmail already filters deleted_at IS NULL.
	if err := h.Users.GetByEmail(req.Email, &user); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := CheckPassword(user.PasswordHash, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.JWTSvc.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		Token: token,
		User:  toUserJSON(&user),
	})
}

// Me handles GET /api/auth/me.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var user models.User
	if err := h.Users.GetByID(claims.UserID, &user); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusUnauthorized, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Reject soft-deleted users even if they have a valid token.
	if user.DeletedAt != nil {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, toUserJSON(&user))
}

// SetupStatus handles GET /api/auth/setup-status.
func (h *Handler) SetupStatus(w http.ResponseWriter, r *http.Request) {
	users, err := h.Users.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"setupRequired": len(users) == 0})
}

// Setup handles POST /api/auth/setup.
func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	// Only allow setup when no users exist.
	users, err := h.Users.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if len(users) > 0 {
		writeError(w, http.StatusConflict, "setup already completed")
		return
	}

	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		Role:         "admin",
		PasswordHash: hash,
	}
	if err := h.Users.Create(&user); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create user: %v", err))
		return
	}

	token, err := h.JWTSvc.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  toUserJSON(&user),
	})
}
