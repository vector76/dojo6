package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/models"
)

// seedPaymentUsers creates an admin and a regular user, returning their IDs and tokens.
func seedPaymentUsers(t *testing.T, h *auth.Handler) (adminID, userID int64, adminToken, userToken string) {
	t.Helper()

	hash, _ := auth.HashPassword("pass")

	admin := models.User{Name: "Admin", Email: "admin@test.com", Phone: "555-0001", Role: "admin", PasswordHash: hash}
	if err := h.Users.Create(&admin); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	adminTok, _ := auth.GenerateToken(testJWTSecret, admin.ID, admin.Email, admin.Role)

	user := models.User{Name: "Student", Email: "student@test.com", Phone: "555-0002", Role: "user", PasswordHash: hash}
	if err := h.Users.Create(&user); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	userTok, _ := auth.GenerateToken(testJWTSecret, user.ID, user.Email, user.Role)

	return admin.ID, user.ID, adminTok, userTok
}

func TestRecordPayment_AdminSuccess(t *testing.T) {
	router, h, _ := testRouter(t)
	_, userID, adminToken, _ := seedPaymentUsers(t, h)

	body := fmt.Sprintf(`{"user_id":%d,"amount":50.00,"date":"2026-02-01","note":"Monthly dues"}`, userID)
	req := httptest.NewRequest(http.MethodPost, "/api/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp paymentJSON
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ID == 0 {
		t.Error("expected non-zero payment ID")
	}
	if resp.Amount != 50.00 {
		t.Errorf("expected amount 50.00, got %f", resp.Amount)
	}
	if resp.UserID != userID {
		t.Errorf("expected user_id %d, got %d", userID, resp.UserID)
	}
	if resp.Note != "Monthly dues" {
		t.Errorf("expected note 'Monthly dues', got '%s'", resp.Note)
	}
}

func TestRecordPayment_NonAdminForbidden(t *testing.T) {
	router, h, _ := testRouter(t)
	_, userID, _, userToken := seedPaymentUsers(t, h)

	body := fmt.Sprintf(`{"user_id":%d,"amount":50.00,"date":"2026-02-01"}`, userID)
	req := httptest.NewRequest(http.MethodPost, "/api/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-admin, got %d", rec.Code)
	}
}

func TestRecordPayment_NoToken(t *testing.T) {
	router, _, _ := testRouter(t)

	body := `{"user_id":1,"amount":50.00,"date":"2026-02-01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", rec.Code)
	}
}

func TestRecordPayment_MissingFields(t *testing.T) {
	router, h, _ := testRouter(t)
	_, _, adminToken, _ := seedPaymentUsers(t, h)

	body := `{"user_id":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing fields, got %d", rec.Code)
	}
}

func TestRecordPayment_NonExistentUser(t *testing.T) {
	router, h, _ := testRouter(t)
	_, _, adminToken, _ := seedPaymentUsers(t, h)

	body := `{"user_id":9999,"amount":50.00,"date":"2026-02-01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent user, got %d", rec.Code)
	}
}

func TestListUserPayments_AdminSuccess(t *testing.T) {
	router, h, ph := testRouter(t)
	adminID, userID, adminToken, _ := seedPaymentUsers(t, h)

	// Create two payments.
	p1 := &models.Payment{UserID: userID, Amount: 50.00, Date: "2026-01-15", RecordedBy: adminID}
	p2 := &models.Payment{UserID: userID, Amount: 75.00, Date: "2026-02-15", RecordedBy: adminID}
	ph.Payments.Create(p1)
	ph.Payments.Create(p2)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/payments", userID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payments []paymentJSON
	json.NewDecoder(rec.Body).Decode(&payments)
	if len(payments) != 2 {
		t.Fatalf("expected 2 payments, got %d", len(payments))
	}
	// Should be ordered by date DESC.
	if payments[0].Date != "2026-02-15" {
		t.Errorf("expected first payment date '2026-02-15', got '%s'", payments[0].Date)
	}
}

func TestListUserPayments_SelfAccess(t *testing.T) {
	router, h, ph := testRouter(t)
	adminID, userID, _, userToken := seedPaymentUsers(t, h)

	p := &models.Payment{UserID: userID, Amount: 40.00, Date: "2026-02-01", RecordedBy: adminID}
	ph.Payments.Create(p)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/payments", userID), nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for self-access, got %d", rec.Code)
	}

	var payments []paymentJSON
	json.NewDecoder(rec.Body).Decode(&payments)
	if len(payments) != 1 {
		t.Errorf("expected 1 payment, got %d", len(payments))
	}
}

func TestListUserPayments_OtherUserForbidden(t *testing.T) {
	router, h, _ := testRouter(t)
	adminID, _, _, userToken := seedPaymentUsers(t, h)

	// Regular user trying to view admin's payments.
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/payments", adminID), nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for accessing other user's payments, got %d", rec.Code)
	}
}

func TestListUserPayments_EmptyList(t *testing.T) {
	router, h, _ := testRouter(t)
	_, userID, adminToken, _ := seedPaymentUsers(t, h)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/payments", userID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payments []paymentJSON
	json.NewDecoder(rec.Body).Decode(&payments)
	if len(payments) != 0 {
		t.Errorf("expected 0 payments, got %d", len(payments))
	}
}

func TestGetBalance_AdminSuccess(t *testing.T) {
	router, h, ph := testRouter(t)
	adminID, userID, adminToken, _ := seedPaymentUsers(t, h)

	// Set expected balance.
	var user models.User
	h.Users.GetByID(userID, &user)
	user.ExpectedBalance = 200.00
	h.Users.Update(&user)

	// Create payments totaling 125.
	p1 := &models.Payment{UserID: userID, Amount: 50.00, Date: "2026-01-15", RecordedBy: adminID}
	p2 := &models.Payment{UserID: userID, Amount: 75.00, Date: "2026-02-15", RecordedBy: adminID}
	ph.Payments.Create(p1)
	ph.Payments.Create(p2)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/balance", userID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp balanceResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ExpectedBalance != 200.00 {
		t.Errorf("expected expected_balance 200.00, got %f", resp.ExpectedBalance)
	}
	if resp.TotalPayments != 125.00 {
		t.Errorf("expected total_payments 125.00, got %f", resp.TotalPayments)
	}
	if resp.Balance != 75.00 {
		t.Errorf("expected balance 75.00, got %f", resp.Balance)
	}
}

func TestGetBalance_SelfAccess(t *testing.T) {
	router, h, _ := testRouter(t)
	_, userID, _, userToken := seedPaymentUsers(t, h)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/balance", userID), nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for self-access, got %d", rec.Code)
	}
}

func TestGetBalance_OtherUserForbidden(t *testing.T) {
	router, h, _ := testRouter(t)
	adminID, _, _, userToken := seedPaymentUsers(t, h)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/balance", adminID), nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for accessing other user's balance, got %d", rec.Code)
	}
}

func TestGetBalance_NonExistentUser(t *testing.T) {
	router, h, _ := testRouter(t)
	_, _, adminToken, _ := seedPaymentUsers(t, h)

	req := httptest.NewRequest(http.MethodGet, "/api/users/9999/balance", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent user, got %d", rec.Code)
	}
}

func TestGetBalance_NoPayments(t *testing.T) {
	router, h, _ := testRouter(t)
	_, userID, adminToken, _ := seedPaymentUsers(t, h)

	// Set expected balance with no payments.
	var user models.User
	h.Users.GetByID(userID, &user)
	user.ExpectedBalance = 100.00
	h.Users.Update(&user)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/balance", userID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp balanceResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ExpectedBalance != 100.00 {
		t.Errorf("expected expected_balance 100.00, got %f", resp.ExpectedBalance)
	}
	if resp.TotalPayments != 0 {
		t.Errorf("expected total_payments 0, got %f", resp.TotalPayments)
	}
	if resp.Balance != 100.00 {
		t.Errorf("expected balance 100.00, got %f", resp.Balance)
	}
}

func TestSetBalance_AdminSuccess(t *testing.T) {
	router, h, _ := testRouter(t)
	_, userID, adminToken, _ := seedPaymentUsers(t, h)

	body := `{"expected_balance":300.00}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/balance", userID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp balanceResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ExpectedBalance != 300.00 {
		t.Errorf("expected expected_balance 300.00, got %f", resp.ExpectedBalance)
	}
	if resp.Balance != 300.00 {
		t.Errorf("expected balance 300.00 (no payments), got %f", resp.Balance)
	}
}

func TestSetBalance_NonAdminForbidden(t *testing.T) {
	router, h, _ := testRouter(t)
	_, userID, _, userToken := seedPaymentUsers(t, h)

	body := `{"expected_balance":300.00}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/balance", userID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-admin setting balance, got %d", rec.Code)
	}
}

func TestSetBalance_NonExistentUser(t *testing.T) {
	router, h, _ := testRouter(t)
	_, _, adminToken, _ := seedPaymentUsers(t, h)

	body := `{"expected_balance":300.00}`
	req := httptest.NewRequest(http.MethodPut, "/api/users/9999/balance", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent user, got %d", rec.Code)
	}
}

func TestSetBalance_ThenRecordPayment_BalanceUpdates(t *testing.T) {
	router, h, _ := testRouter(t)
	_, userID, adminToken, _ := seedPaymentUsers(t, h)

	// 1. Set expected balance to 200.
	setBody := `{"expected_balance":200.00}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/balance", userID), strings.NewReader(setBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("set balance: expected 200, got %d", rec.Code)
	}

	// 2. Record a payment of 80.
	payBody := fmt.Sprintf(`{"user_id":%d,"amount":80.00,"date":"2026-02-15"}`, userID)
	req = httptest.NewRequest(http.MethodPost, "/api/payments", strings.NewReader(payBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("record payment: expected 201, got %d", rec.Code)
	}

	// 3. Check balance: expected 200 - 80 = 120.
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/balance", userID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("get balance: expected 200, got %d", rec.Code)
	}

	var resp balanceResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ExpectedBalance != 200.00 {
		t.Errorf("expected expected_balance 200.00, got %f", resp.ExpectedBalance)
	}
	if resp.TotalPayments != 80.00 {
		t.Errorf("expected total_payments 80.00, got %f", resp.TotalPayments)
	}
	if resp.Balance != 120.00 {
		t.Errorf("expected balance 120.00, got %f", resp.Balance)
	}
}

func TestRecordPayment_InstructorForbidden(t *testing.T) {
	router, h, _ := testRouter(t)
	_, userID, _, _ := seedPaymentUsers(t, h)

	// Create instructor.
	hash, _ := auth.HashPassword("pass")
	instructor := models.User{Name: "Inst", Email: "inst@test.com", Phone: "555-0003", Role: "instructor", PasswordHash: hash}
	h.Users.Create(&instructor)
	instToken, _ := auth.GenerateToken(testJWTSecret, instructor.ID, instructor.Email, instructor.Role)

	body := fmt.Sprintf(`{"user_id":%d,"amount":50.00,"date":"2026-02-01"}`, userID)
	req := httptest.NewRequest(http.MethodPost, "/api/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+instToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for instructor, got %d", rec.Code)
	}
}

func TestSetBalance_InstructorForbidden(t *testing.T) {
	router, h, _ := testRouter(t)
	_, userID, _, _ := seedPaymentUsers(t, h)

	hash, _ := auth.HashPassword("pass")
	instructor := models.User{Name: "Inst", Email: "inst@test.com", Phone: "555-0003", Role: "instructor", PasswordHash: hash}
	h.Users.Create(&instructor)
	instToken, _ := auth.GenerateToken(testJWTSecret, instructor.ID, instructor.Email, instructor.Role)

	body := `{"expected_balance":300.00}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/balance", userID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+instToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for instructor setting balance, got %d", rec.Code)
	}
}
