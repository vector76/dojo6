package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey contextKey = "user"

// UserContext holds the authenticated user's information extracted from the JWT.
type UserContext struct {
	UserID int64
	Email  string
	Role   string
}

// JWTMiddleware returns an HTTP middleware that validates the Authorization
// header (Bearer token) and injects a UserContext into the request context.
func JWTMiddleware(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, "authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString, found := strings.CutPrefix(authHeader, "Bearer ")
			if !found {
				writeJSONError(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				writeJSONError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			uc := &UserContext{
				UserID: claims.UserID,
				Email:  claims.Email,
				Role:   claims.Role,
			}

			ctx := context.WithValue(r.Context(), userContextKey, uc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUser extracts the UserContext from the request context.
// Returns nil if no user context is present.
func GetUser(ctx context.Context) *UserContext {
	uc, _ := ctx.Value(userContextKey).(*UserContext)
	return uc
}

func writeJSONError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}
