package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// PaymentHandler holds dependencies for payment HTTP handlers.
type PaymentHandler struct {
	Payments *models.PaymentRepository
	Users    *models.UserRepository
}

type createPaymentRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
	Note   string  `json:"note"`
}

type paymentJSON struct {
	ID         int64   `json:"id"`
	UserID     int64   `json:"user_id"`
	Amount     float64 `json:"amount"`
	Date       string  `json:"date"`
	Note       string  `json:"note"`
	RecordedBy int64   `json:"recorded_by"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type balanceResponse struct {
	ExpectedBalance float64 `json:"expected_balance"`
	TotalPayments   float64 `json:"total_payments"`
	Balance         float64 `json:"balance"`
}

type setBalanceRequest struct {
	ExpectedBalance float64 `json:"expected_balance"`
}

func toPaymentJSON(p *models.Payment) paymentJSON {
	return paymentJSON{
		ID:         p.ID,
		UserID:     p.UserID,
		Amount:     p.Amount,
		Date:       p.Date,
		Note:       p.Note,
		RecordedBy: p.RecordedBy,
		CreatedAt:  p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// RecordPayment handles POST /api/payments (admin only).
func (h *PaymentHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == 0 || req.Amount == 0 || req.Date == "" {
		writeError(w, http.StatusBadRequest, "user_id, amount, and date are required")
		return
	}

	// Verify the target user exists.
	var user models.User
	if err := h.Users.GetByID(req.UserID, &user); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	p := models.Payment{
		UserID:     req.UserID,
		Amount:     req.Amount,
		Date:       req.Date,
		Note:       req.Note,
		RecordedBy: claims.UserID,
	}
	if err := h.Payments.Create(&p); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record payment")
		return
	}

	writeJSON(w, http.StatusCreated, toPaymentJSON(&p))
}

// ListUserPayments handles GET /api/users/:id/payments (admin or self).
func (h *PaymentHandler) ListUserPayments(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if claims.Role != "admin" && claims.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	payments, err := h.Payments.ListByUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	result := make([]paymentJSON, len(payments))
	for i := range payments {
		result[i] = toPaymentJSON(&payments[i])
	}
	writeJSON(w, http.StatusOK, result)
}

// GetBalance handles GET /api/users/:id/balance (admin or self).
func (h *PaymentHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if claims.Role != "admin" && claims.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var user models.User
	if err := h.Users.GetByID(userID, &user); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	payments, err := h.Payments.ListByUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var totalPayments float64
	for _, p := range payments {
		totalPayments += p.Amount
	}

	writeJSON(w, http.StatusOK, balanceResponse{
		ExpectedBalance: user.ExpectedBalance,
		TotalPayments:   totalPayments,
		Balance:         user.ExpectedBalance - totalPayments,
	})
}

// SetBalance handles PUT /api/users/:id/balance (admin only).
func (h *PaymentHandler) SetBalance(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req setBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var user models.User
	if err := h.Users.GetByID(userID, &user); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	user.ExpectedBalance = req.ExpectedBalance
	if err := h.Users.Update(&user); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update balance")
		return
	}

	// Return the updated balance info.
	payments, err := h.Payments.ListByUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var totalPayments float64
	for _, p := range payments {
		totalPayments += p.Amount
	}

	writeJSON(w, http.StatusOK, balanceResponse{
		ExpectedBalance: user.ExpectedBalance,
		TotalPayments:   totalPayments,
		Balance:         user.ExpectedBalance - totalPayments,
	})
}
