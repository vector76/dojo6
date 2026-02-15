package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const claimsKey contextKey = "claims"

// UserContext holds the authenticated user's information extracted from the JWT.
type UserContext struct {
	UserID int64
	Email  string
	Role   string
}

// JWTMiddleware validates the Authorization header and injects claims into
// the request context. Returns 401 for missing or invalid tokens.
func JWTMiddleware(svc *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			tokenStr, found := strings.CutPrefix(header, "Bearer ")
			if !found {
				writeJSONError(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, err := svc.ValidateToken(tokenStr)
			if err != nil {
				writeJSONError(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims retrieves the JWT claims from the request context.
func GetClaims(r *http.Request) *Claims {
	claims, _ := r.Context().Value(claimsKey).(*Claims)
	return claims
}

// GetUser extracts a UserContext from the request context.
// Returns nil if no claims are present.
func GetUser(ctx context.Context) *UserContext {
	claims, _ := ctx.Value(claimsKey).(*Claims)
	if claims == nil {
		return nil
	}
	return &UserContext{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}
}

// RequireRole returns middleware that checks whether the authenticated user
// has one of the allowed roles. Returns 403 if not.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r)
			if claims == nil || !allowed[claims.Role] {
				writeJSONError(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeJSONError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + msg + `"}`))
}
