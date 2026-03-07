package api

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

type Event struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	EventType     string `json:"event_type"`
	Status        string `json:"status"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Location      string `json:"location"`
	OrganizerID   int64  `json:"organizer_id"`
	OrganizerName string `json:"organizer_name"`
}

type EventList struct {
	Events []Event `json:"events"`
	Total  int     `json:"total"`
}

type CreateEventRequest struct {
	Title          string  `json:"title"`
	Description    string  `json:"description,omitempty"`
	EventType      string  `json:"event_type"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	Location       string  `json:"location,omitempty"`
	ParticipantIDs []int64 `json:"participant_ids,omitempty"`
	Priority       string  `json:"priority,omitempty"`
}

// CreateEvent creates a new schedule event.
func (c *Client) CreateEvent(ctx context.Context, a *agent.Agent, req CreateEventRequest) (*Event, error) {
	resp, err := c.Post(ctx, "/api/events", a, req)
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	var event Event
	if err := ParseData(resp, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// ListEvents retrieves events.
func (c *Client) ListEvents(ctx context.Context, a *agent.Agent, queryParams string) (*EventList, error) {
	path := "/api/events"
	if queryParams != "" {
		path += "?" + queryParams
	}
	resp, err := c.Get(ctx, path, a)
	if err != nil {
		return nil, err
	}
	var list EventList
	if err := ParseData(resp, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetUpcomingEvents retrieves upcoming events.
func (c *Client) GetUpcomingEvents(ctx context.Context, a *agent.Agent) (*EventList, error) {
	resp, err := c.Get(ctx, "/api/events/upcoming", a)
	if err != nil {
		return nil, err
	}
	var list EventList
	if err := ParseData(resp, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// AddEventParticipants adds participants to an event.
func (c *Client) AddEventParticipants(ctx context.Context, a *agent.Agent, eventID int64, userIDs []int64) error {
	body := map[string]any{
		"user_ids": userIDs,
	}
	_, err := c.Post(ctx, fmt.Sprintf("/api/events/%d/participants", eventID), a, body)
	return err
}

// RespondToEvent responds to an event invitation.
func (c *Client) RespondToEvent(ctx context.Context, a *agent.Agent, eventID int64, status string) error {
	body := map[string]any{
		"status": status,
	}
	_, err := c.Post(ctx, fmt.Sprintf("/api/events/%d/respond", eventID), a, body)
	return err
}
