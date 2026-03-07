package api

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	User         authUser `json:"user"`
}

type authUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// Register creates a new user account for the agent.
func (c *Client) Register(ctx context.Context, a *agent.Agent) error {
	req := registerRequest{
		Name:     a.Name,
		Email:    a.Email,
		Password: a.Password,
		Role:     a.Role,
	}

	resp, err := c.Post(ctx, "/api/auth/register", nil, req)
	if err != nil {
		return fmt.Errorf("register %s: %w", a.Email, err)
	}

	var auth authResponse
	if err := ParseData(resp, &auth); err != nil {
		return fmt.Errorf("parse register response for %s: %w", a.Email, err)
	}

	a.AccessToken = auth.Token
	a.RefreshToken = auth.RefreshToken
	a.UserID = auth.User.ID

	return nil
}

// Login authenticates the agent and stores tokens.
func (c *Client) Login(ctx context.Context, a *agent.Agent) error {
	req := loginRequest{
		Email:    a.Email,
		Password: a.Password,
	}

	resp, err := c.Post(ctx, "/api/auth/login", nil, req)
	if err != nil {
		return fmt.Errorf("login %s: %w", a.Email, err)
	}

	var auth authResponse
	if err := ParseData(resp, &auth); err != nil {
		return fmt.Errorf("parse login response for %s: %w", a.Email, err)
	}

	a.AccessToken = auth.Token
	a.RefreshToken = auth.RefreshToken
	a.UserID = auth.User.ID

	return nil
}

// RefreshToken refreshes the agent's access token.
func (c *Client) RefreshToken(ctx context.Context, a *agent.Agent) error {
	req := refreshRequest{
		RefreshToken: a.RefreshToken,
	}

	resp, err := c.Post(ctx, "/api/auth/refresh", nil, req)
	if err != nil {
		return fmt.Errorf("refresh token for %s: %w", a.Email, err)
	}

	var auth authResponse
	if err := ParseData(resp, &auth); err != nil {
		return fmt.Errorf("parse refresh response for %s: %w", a.Email, err)
	}

	a.AccessToken = auth.Token
	if auth.RefreshToken != "" {
		a.RefreshToken = auth.RefreshToken
	}

	return nil
}

// EnsureAuthenticated logs in the agent if not already authenticated.
func (c *Client) EnsureAuthenticated(ctx context.Context, a *agent.Agent) error {
	if a.IsAuthenticated() {
		return nil
	}
	return c.Login(ctx, a)
}
