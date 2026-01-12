package jwt

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	// ClaimsContextKey is the context key for JWT claims
	ClaimsContextKey contextKey = "jwt_claims"
)

// Middleware returns an HTTP middleware that validates JWT tokens
func (m *JWTManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := m.ValidateToken(tokenString)
		if err != nil {
			if err == ErrExpiredToken {
				http.Error(w, "token has expired", http.StatusUnauthorized)
			} else {
				http.Error(w, "invalid token", http.StatusUnauthorized)
			}
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetClaimsFromContext extracts JWT claims from context
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}
