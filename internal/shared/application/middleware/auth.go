// Package middleware contains shared HTTP middleware components.
package middleware

import (
	"net/http"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	// TODO: Add dependencies (token service, etc.)
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// Handle authenticates incoming requests
func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement authentication logic
		// 1. Extract token from header
		// 2. Validate token
		// 3. Add user info to context
		// 4. Call next handler

		next.ServeHTTP(w, r)
	})
}
