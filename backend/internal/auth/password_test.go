package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{name: "simple password", password: "password123"},
		{name: "complex password", password: "P@$$w0rd!#%^&*()"},
		{name: "empty password", password: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if hash == "" {
				t.Error("expected non-empty hash")
			}
			if hash == tt.password {
				t.Error("hash should not equal plaintext password")
			}
		})
	}
}

func TestHashPassword_ProducesDifferentHashes(t *testing.T) {
	hash1, err := HashPassword("same-password")
	if err != nil {
		t.Fatalf("hash1: %v", err)
	}
	hash2, err := HashPassword("same-password")
	if err != nil {
		t.Fatalf("hash2: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected different hashes for same password (bcrypt uses random salt)")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "correct-password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "correct password", password: "correct-password"},
		{name: "wrong password", password: "wrong-password", wantErr: true},
		{name: "empty password", password: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(hash, tt.password)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	err := CheckPassword("not-a-valid-hash", "password")
	if err == nil {
		t.Error("expected error for invalid hash")
	}
}

func TestHashPassword_UsesBcrypt(t *testing.T) {
	hash, err := HashPassword("test")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	// Verify it's a valid bcrypt hash by checking cost.
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("bcrypt.Cost: %v", err)
	}
	if cost != bcrypt.DefaultCost {
		t.Errorf("expected cost %d, got %d", bcrypt.DefaultCost, cost)
	}
}
