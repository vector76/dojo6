package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/models"
)

func TestSetupStatus_NoUsers(t *testing.T) {
	router, _ := testRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/setup-status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]bool
	json.NewDecoder(rec.Body).Decode(&body)
	if !body["setupRequired"] {
		t.Error("expected setupRequired=true when no users exist")
	}
}

func TestSetup_CreatesAdmin(t *testing.T) {
	router, _ := testRouter(t)

	body := `{"email":"admin@dojo.com","password":"secret123","name":"Admin","phone":"555-0100"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Token string `json:"token"`
		User  struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		} `json:"user"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Email != "admin@dojo.com" {
		t.Errorf("expected email admin@dojo.com, got %s", resp.User.Email)
	}
	if resp.User.Role != "admin" {
		t.Errorf("expected role admin, got %s", resp.User.Role)
	}
	if resp.User.Name != "Admin" {
		t.Errorf("expected name Admin, got %s", resp.User.Name)
	}
}

func TestSetup_FailsWhenUsersExist(t *testing.T) {
	router, h := testRouter(t)

	// Create a user first.
	hash, _ := auth.HashPassword("pw")
	user := models.User{Name: "A", Email: "a@b.com", Role: "admin", PasswordHash: hash}
	h.Users.Create(&user)

	body := `{"email":"new@dojo.com","password":"secret","name":"New"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
}

func TestSetupStatus_AfterSetup(t *testing.T) {
	router, h := testRouter(t)

	hash, _ := auth.HashPassword("pw")
	user := models.User{Name: "A", Email: "a@b.com", Role: "admin", PasswordHash: hash}
	h.Users.Create(&user)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/setup-status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var body map[string]bool
	json.NewDecoder(rec.Body).Decode(&body)
	if body["setupRequired"] {
		t.Error("expected setupRequired=false when users exist")
	}
}

func TestLogin_Success(t *testing.T) {
	router, h := testRouter(t)

	hash, _ := auth.HashPassword("mypassword")
	user := models.User{Name: "Alice", Email: "alice@dojo.com", Phone: "555-0101", Role: "user", PasswordHash: hash}
	h.Users.Create(&user)

	body := `{"email":"alice@dojo.com","password":"mypassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Token string `json:"token"`
		User  struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Email != "alice@dojo.com" {
		t.Errorf("expected email alice@dojo.com, got %s", resp.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	router, h := testRouter(t)

	hash, _ := auth.HashPassword("correct")
	user := models.User{Name: "A", Email: "a@b.com", Role: "user", PasswordHash: hash}
	h.Users.Create(&user)

	body := `{"email":"a@b.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestLogin_NonexistentEmail(t *testing.T) {
	router, _ := testRouter(t)

	body := `{"email":"nobody@dojo.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestLogin_SoftDeletedUser(t *testing.T) {
	router, h := testRouter(t)

	hash, _ := auth.HashPassword("pass")
	user := models.User{Name: "Del", Email: "del@b.com", Role: "user", PasswordHash: hash}
	h.Users.Create(&user)
	h.Users.SoftDelete(user.ID)

	body := `{"email":"del@b.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for soft-deleted user, got %d", rec.Code)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	router, _ := testRouter(t)

	body := `{"email":"a@b.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMe_WithValidToken(t *testing.T) {
	router, h := testRouter(t)

	hash, _ := auth.HashPassword("pass")
	user := models.User{Name: "Alice", Email: "alice@dojo.com", Phone: "555-0101", Role: "instructor", PasswordHash: hash}
	h.Users.Create(&user)

	token, _ := auth.GenerateToken(testJWTSecret, user.ID, user.Email, user.Role)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Email string `json:"email"`
		Role  string `json:"role"`
		Name  string `json:"name"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Email != "alice@dojo.com" {
		t.Errorf("expected email alice@dojo.com, got %s", resp.Email)
	}
	if resp.Role != "instructor" {
		t.Errorf("expected role instructor, got %s", resp.Role)
	}
}

func TestMe_WithoutToken(t *testing.T) {
	router, _ := testRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMe_WithInvalidToken(t *testing.T) {
	router, _ := testRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMe_SoftDeletedUser(t *testing.T) {
	router, h := testRouter(t)

	hash, _ := auth.HashPassword("pass")
	user := models.User{Name: "Del", Email: "del@b.com", Role: "user", PasswordHash: hash}
	h.Users.Create(&user)

	// Generate token before deleting.
	token, _ := auth.GenerateToken(testJWTSecret, user.ID, user.Email, user.Role)
	h.Users.SoftDelete(user.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for deleted user, got %d", rec.Code)
	}
}

func TestAuthRoundTrip(t *testing.T) {
	router, _ := testRouter(t)

	// 1. Setup creates admin.
	setupBody := `{"email":"admin@dojo.com","password":"admin123","name":"Admin","phone":"555-0100"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", strings.NewReader(setupBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("setup: expected 201, got %d", rec.Code)
	}

	var setupResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(rec.Body).Decode(&setupResp)

	// 2. Use token from setup to access /me.
	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+setupResp.Token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("me after setup: expected 200, got %d", rec.Code)
	}

	// 3. Login with same credentials.
	loginBody := `{"email":"admin@dojo.com","password":"admin123"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d", rec.Code)
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(rec.Body).Decode(&loginResp)

	// 4. Use login token to access /me.
	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("me after login: expected 200, got %d", rec.Code)
	}

	// 5. Setup is now locked.
	req = httptest.NewRequest(http.MethodPost, "/api/auth/setup", strings.NewReader(setupBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("setup after admin exists: expected 409, got %d", rec.Code)
	}
}

func TestRoleEnforcement(t *testing.T) {
	router, h := testRouter(t)

	// Create users with different roles.
	roles := map[string]*models.User{
		"admin":      {Name: "Admin", Email: "admin@d.com", Role: "admin"},
		"instructor": {Name: "Inst", Email: "inst@d.com", Role: "instructor"},
		"user":       {Name: "User", Email: "user@d.com", Role: "user"},
	}
	tokens := make(map[string]string)

	for role, u := range roles {
		hash, _ := auth.HashPassword("pass")
		u.PasswordHash = hash
		h.Users.Create(u)
		tok, _ := auth.GenerateToken(testJWTSecret, u.ID, u.Email, u.Role)
		tokens[role] = tok
		_ = role
	}

	// All roles should be able to access /me.
	for role, tok := range tokens {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("%s: expected 200 on /me, got %d", role, rec.Code)
		}
	}
}
