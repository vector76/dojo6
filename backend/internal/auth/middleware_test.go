package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestJWTMiddleware_ValidToken(t *testing.T) {
	svc := newTestJWTService()
	token, err := svc.GenerateToken(42, "alice@example.com", "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var capturedUser *UserContext
	handler := JWTMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedUser == nil {
		t.Fatal("expected user context to be set")
	}
	if capturedUser.UserID != 42 {
		t.Errorf("expected userID 42, got %d", capturedUser.UserID)
	}
	if capturedUser.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", capturedUser.Email)
	}
	if capturedUser.Role != "admin" {
		t.Errorf("expected role admin, got %s", capturedUser.Role)
	}
}

func TestJWTMiddleware_MissingHeader(t *testing.T) {
	svc := newTestJWTService()
	handler := JWTMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}

func TestJWTMiddleware_InvalidFormat(t *testing.T) {
	svc := newTestJWTService()
	handler := JWTMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	tests := []struct {
		name   string
		header string
	}{
		{name: "no Bearer prefix", header: "Token abc123"},
		{name: "Basic auth", header: "Basic dXNlcjpwYXNz"},
		{name: "just Bearer", header: "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	svc := NewJWTService("test-secret", -time.Second)
	token, err := svc.GenerateToken(1, "alice@example.com", "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	handler := JWTMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	svc := newTestJWTService()
	handler := JWTMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestGetUser_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	uc := GetUser(req.Context())
	if uc != nil {
		t.Error("expected nil UserContext when not set")
	}
}
