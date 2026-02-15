package auth

import (
	"testing"
	"time"
)

func newTestJWTService() *JWTService {
	return NewJWTService("test-secret-key", time.Hour)
}

func TestGenerateToken(t *testing.T) {
	svc := newTestJWTService()

	token, err := svc.GenerateToken(1, "alice@example.com", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestValidateToken_RoundTrip(t *testing.T) {
	svc := newTestJWTService()

	tests := []struct {
		name   string
		userID int64
		email  string
		role   string
	}{
		{name: "admin", userID: 1, email: "admin@example.com", role: "admin"},
		{name: "instructor", userID: 2, email: "instructor@example.com", role: "instructor"},
		{name: "user", userID: 3, email: "user@example.com", role: "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := svc.GenerateToken(tt.userID, tt.email, tt.role)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}

			claims, err := svc.ValidateToken(token)
			if err != nil {
				t.Fatalf("validate: %v", err)
			}

			if claims.UserID != tt.userID {
				t.Errorf("expected userID %d, got %d", tt.userID, claims.UserID)
			}
			if claims.Email != tt.email {
				t.Errorf("expected email %s, got %s", tt.email, claims.Email)
			}
			if claims.Role != tt.role {
				t.Errorf("expected role %s, got %s", tt.role, claims.Role)
			}
		})
	}
}

func TestValidateToken_Expired(t *testing.T) {
	// Service with zero duration produces already-expired tokens.
	svc := NewJWTService("test-secret", -time.Second)

	token, err := svc.GenerateToken(1, "alice@example.com", "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = svc.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateToken_InvalidString(t *testing.T) {
	svc := newTestJWTService()

	tests := []struct {
		name  string
		token string
	}{
		{name: "empty string", token: ""},
		{name: "garbage", token: "not.a.token"},
		{name: "tampered", token: "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoxfQ.tampered"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ValidateToken(tt.token)
			if err != ErrInvalidToken {
				t.Errorf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-one", time.Hour)
	svc2 := NewJWTService("secret-two", time.Hour)

	token, err := svc1.GenerateToken(1, "alice@example.com", "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = svc2.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for wrong secret, got %v", err)
	}
}
