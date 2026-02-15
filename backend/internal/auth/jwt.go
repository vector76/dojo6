package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// Claims holds the JWT payload for an authenticated user.
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for the given user.
func GenerateToken(secret string, userID int64, email, role string) (string, error) {
	return GenerateTokenWithDuration(secret, userID, email, role, 24*time.Hour)
}

// GenerateTokenWithDuration creates a signed JWT with a custom duration.
func GenerateTokenWithDuration(secret string, userID int64, email, role string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and validates a JWT string, returning the claims.
func ValidateToken(secret, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// JWTService handles token issuance and validation.
type JWTService struct {
	secret   string
	duration time.Duration
}

// NewJWTService creates a new JWT service with the given secret and token duration.
func NewJWTService(secret string, duration time.Duration) *JWTService {
	return &JWTService{
		secret:   secret,
		duration: duration,
	}
}

// GenerateToken creates a signed JWT for the given user.
func (s *JWTService) GenerateToken(userID int64, email, role string) (string, error) {
	return GenerateTokenWithDuration(s.secret, userID, email, role, s.duration)
}

// ValidateToken parses and validates a JWT string, returning the claims.
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	return ValidateToken(s.secret, tokenString)
}
