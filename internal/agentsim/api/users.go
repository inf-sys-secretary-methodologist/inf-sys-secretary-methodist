package api

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

// User represents a system user.
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// UserList represents a list of users.
type UserList struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}

// ListUsers retrieves the list of users.
func (c *Client) ListUsers(ctx context.Context, a *agent.Agent) (*UserList, error) {
	resp, err := c.Get(ctx, "/api/users", a)
	if err != nil {
		return nil, err
	}
	var list UserList
	if err := ParseData(resp, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetCurrentUser retrieves the currently authenticated user's info.
func (c *Client) GetCurrentUser(ctx context.Context, a *agent.Agent) (*User, error) {
	resp, err := c.Get(ctx, "/api/me", a)
	if err != nil {
		return nil, err
	}
	var user User
	if err := ParseData(resp, &user); err != nil {
		return nil, err
	}
	return &user, nil
}
