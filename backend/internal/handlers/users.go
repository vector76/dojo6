package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/models"
)

// validRoles defines the allowed role values.
var validRoles = map[string]bool{
	"admin":      true,
	"instructor": true,
	"user":       true,
}

// UserHandler holds dependencies for user-related HTTP handlers.
type UserHandler struct {
	repo *models.UserRepository
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(repo *models.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// userResponse is the JSON representation of a user, excluding sensitive fields.
type userResponse struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Email            string  `json:"email"`
	Phone            string  `json:"phone"`
	Role             string  `json:"role"`
	MembershipType   string  `json:"membership_type"`
	MembershipStatus string  `json:"membership_status"`
	EmergencyContact string  `json:"emergency_contact"`
	JoinDate         string  `json:"join_date"`
	ExpectedBalance  float64 `json:"expected_balance"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func toUserResponse(u *models.User) userResponse {
	return userResponse{
		ID:               u.ID,
		Name:             u.Name,
		Email:            u.Email,
		Phone:            u.Phone,
		Role:             u.Role,
		MembershipType:   u.MembershipType,
		MembershipStatus: u.MembershipStatus,
		EmergencyContact: u.EmergencyContact,
		JoinDate:         u.JoinDate,
		ExpectedBalance:  u.ExpectedBalance,
		CreatedAt:        u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// List handles GET /api/users (admin and instructor only).
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	uc := auth.GetUser(r.Context())
	if uc == nil {
		auth.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if uc.Role != "admin" && uc.Role != "instructor" {
		auth.WriteJSONError(w, "forbidden", http.StatusForbidden)
		return
	}

	users, err := h.repo.List()
	if err != nil {
		auth.WriteJSONError(w, "failed to list users", http.StatusInternalServerError)
		return
	}

	result := make([]userResponse, len(users))
	for i := range users {
		result[i] = toUserResponse(&users[i])
	}
	writeJSON(w, http.StatusOK, result)
}

// Create handles POST /api/users (admin only).
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	uc := auth.GetUser(r.Context())
	if uc == nil {
		auth.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if uc.Role != "admin" {
		auth.WriteJSONError(w, "forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Name             string  `json:"name"`
		Email            string  `json:"email"`
		Phone            string  `json:"phone"`
		Role             string  `json:"role"`
		Password         string  `json:"password"`
		MembershipType   string  `json:"membership_type"`
		MembershipStatus string  `json:"membership_status"`
		EmergencyContact string  `json:"emergency_contact"`
		JoinDate         string  `json:"join_date"`
		ExpectedBalance  float64 `json:"expected_balance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields.
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Password = strings.TrimSpace(req.Password)

	if req.Name == "" {
		auth.WriteJSONError(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		auth.WriteJSONError(w, "email is required", http.StatusBadRequest)
		return
	}
	if req.Phone == "" {
		auth.WriteJSONError(w, "phone is required", http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		auth.WriteJSONError(w, "password is required", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}
	if !validRoles[req.Role] {
		auth.WriteJSONError(w, "invalid role: must be admin, instructor, or user", http.StatusBadRequest)
		return
	}

	// Check email uniqueness.
	var existing models.User
	err := h.repo.GetByEmail(req.Email, &existing)
	if err == nil {
		auth.WriteJSONError(w, "email already exists", http.StatusConflict)
		return
	}
	if err != sql.ErrNoRows {
		auth.WriteJSONError(w, "failed to check email", http.StatusInternalServerError)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		auth.WriteJSONError(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	u := &models.User{
		Name:             req.Name,
		Email:            req.Email,
		Phone:            req.Phone,
		Role:             req.Role,
		PasswordHash:     hash,
		MembershipType:   req.MembershipType,
		MembershipStatus: req.MembershipStatus,
		EmergencyContact: req.EmergencyContact,
		JoinDate:         req.JoinDate,
		ExpectedBalance:  req.ExpectedBalance,
	}

	if err := h.repo.Create(u); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			auth.WriteJSONError(w, "email already exists", http.StatusConflict)
			return
		}
		auth.WriteJSONError(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(u))
}

// GetByID handles GET /api/users/{id} (admin, instructor, or self).
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	uc := auth.GetUser(r.Context())
	if uc == nil {
		auth.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		auth.WriteJSONError(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	// Admin and instructor can view any user; regular users can only view themselves.
	if uc.Role != "admin" && uc.Role != "instructor" && uc.UserID != id {
		auth.WriteJSONError(w, "forbidden", http.StatusForbidden)
		return
	}

	var u models.User
	if err := h.repo.GetByID(id, &u); err != nil {
		if err == sql.ErrNoRows {
			auth.WriteJSONError(w, "user not found", http.StatusNotFound)
			return
		}
		auth.WriteJSONError(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	// Don't expose soft-deleted users.
	if u.DeletedAt != nil {
		auth.WriteJSONError(w, "user not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(&u))
}

// Update handles PUT /api/users/{id} (admin or self).
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	uc := auth.GetUser(r.Context())
	if uc == nil {
		auth.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		auth.WriteJSONError(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	if uc.Role != "admin" && uc.UserID != id {
		auth.WriteJSONError(w, "forbidden", http.StatusForbidden)
		return
	}

	var u models.User
	if err := h.repo.GetByID(id, &u); err != nil {
		if err == sql.ErrNoRows {
			auth.WriteJSONError(w, "user not found", http.StatusNotFound)
			return
		}
		auth.WriteJSONError(w, "failed to get user", http.StatusInternalServerError)
		return
	}
	if u.DeletedAt != nil {
		auth.WriteJSONError(w, "user not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name             *string  `json:"name"`
		Email            *string  `json:"email"`
		Phone            *string  `json:"phone"`
		MembershipType   *string  `json:"membership_type"`
		MembershipStatus *string  `json:"membership_status"`
		EmergencyContact *string  `json:"emergency_contact"`
		JoinDate         *string  `json:"join_date"`
		ExpectedBalance  *float64 `json:"expected_balance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			auth.WriteJSONError(w, "name cannot be empty", http.StatusBadRequest)
			return
		}
		u.Name = name
	}
	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email == "" {
			auth.WriteJSONError(w, "email cannot be empty", http.StatusBadRequest)
			return
		}
		// Check uniqueness if email is changing.
		if email != u.Email {
			var existing models.User
			lookupErr := h.repo.GetByEmail(email, &existing)
			if lookupErr == nil {
				auth.WriteJSONError(w, "email already exists", http.StatusConflict)
				return
			}
			if lookupErr != sql.ErrNoRows {
				auth.WriteJSONError(w, "failed to check email", http.StatusInternalServerError)
				return
			}
		}
		u.Email = email
	}
	if req.Phone != nil {
		u.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.MembershipType != nil {
		u.MembershipType = *req.MembershipType
	}
	if req.MembershipStatus != nil {
		u.MembershipStatus = *req.MembershipStatus
	}
	if req.EmergencyContact != nil {
		u.EmergencyContact = *req.EmergencyContact
	}
	if req.JoinDate != nil {
		u.JoinDate = *req.JoinDate
	}
	if req.ExpectedBalance != nil {
		u.ExpectedBalance = *req.ExpectedBalance
	}

	if err := h.repo.Update(&u); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			auth.WriteJSONError(w, "email already exists", http.StatusConflict)
			return
		}
		auth.WriteJSONError(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(&u))
}

// Delete handles DELETE /api/users/{id} (admin only).
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uc := auth.GetUser(r.Context())
	if uc == nil {
		auth.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if uc.Role != "admin" {
		auth.WriteJSONError(w, "forbidden", http.StatusForbidden)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		auth.WriteJSONError(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.SoftDelete(id); err != nil {
		if err == sql.ErrNoRows {
			auth.WriteJSONError(w, "user not found", http.StatusNotFound)
			return
		}
		auth.WriteJSONError(w, "failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ChangeRole handles PUT /api/users/{id}/role (admin only).
func (h *UserHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	uc := auth.GetUser(r.Context())
	if uc == nil {
		auth.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if uc.Role != "admin" {
		auth.WriteJSONError(w, "forbidden", http.StatusForbidden)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		auth.WriteJSONError(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if !validRoles[req.Role] {
		auth.WriteJSONError(w, "invalid role: must be admin, instructor, or user", http.StatusBadRequest)
		return
	}

	var u models.User
	if err := h.repo.GetByID(id, &u); err != nil {
		if err == sql.ErrNoRows {
			auth.WriteJSONError(w, "user not found", http.StatusNotFound)
			return
		}
		auth.WriteJSONError(w, "failed to get user", http.StatusInternalServerError)
		return
	}
	if u.DeletedAt != nil {
		auth.WriteJSONError(w, "user not found", http.StatusNotFound)
		return
	}

	u.Role = req.Role
	if err := h.repo.Update(&u); err != nil {
		auth.WriteJSONError(w, "failed to update role", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(&u))
}

// ChangePassword handles PUT /api/users/{id}/password (admin only).
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	uc := auth.GetUser(r.Context())
	if uc == nil {
		auth.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if uc.Role != "admin" {
		auth.WriteJSONError(w, "forbidden", http.StatusForbidden)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		auth.WriteJSONError(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Password = strings.TrimSpace(req.Password)
	if req.Password == "" {
		auth.WriteJSONError(w, "password is required", http.StatusBadRequest)
		return
	}

	var u models.User
	if err := h.repo.GetByID(id, &u); err != nil {
		if err == sql.ErrNoRows {
			auth.WriteJSONError(w, "user not found", http.StatusNotFound)
			return
		}
		auth.WriteJSONError(w, "failed to get user", http.StatusInternalServerError)
		return
	}
	if u.DeletedAt != nil {
		auth.WriteJSONError(w, "user not found", http.StatusNotFound)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		auth.WriteJSONError(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	u.PasswordHash = hash
	if err := h.repo.Update(&u); err != nil {
		auth.WriteJSONError(w, "failed to update password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
