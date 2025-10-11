package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	loginUseCase *usecases.LoginUserUseCase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(loginUseCase *usecases.LoginUserUseCase) *AuthHandler {
	return &AuthHandler{
		loginUseCase: loginUseCase,
	}
}

// LoginRequest represents login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents login response body
type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// UserDTO represents user data transfer object
type UserDTO struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute use case
	result, err := h.loginUseCase.Execute(usecases.LoginUserCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Build response
	resp := LoginResponse{
		Token: result.Token,
		User: UserDTO{
			ID:     result.User.ID,
			Email:  result.User.Email,
			Name:   result.User.Name,
			Role:   string(result.User.Role),
			Status: string(result.User.Status),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp) //nolint:errcheck // response encoding
}
