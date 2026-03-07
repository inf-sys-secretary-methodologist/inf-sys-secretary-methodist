package api

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

type Notification struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Priority string `json:"priority"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	IsRead   bool   `json:"is_read"`
	Link     string `json:"link"`
}

type NotificationList struct {
	Notifications []Notification `json:"notifications"`
	TotalCount    int            `json:"total_count"`
	UnreadCount   int            `json:"unread_count"`
}

type UnreadCount struct {
	Count int `json:"count"`
}

// ListNotifications retrieves notifications for the agent.
func (c *Client) ListNotifications(ctx context.Context, a *agent.Agent, limit int) (*NotificationList, error) {
	resp, err := c.Get(ctx, "/api/notifications?limit="+itoa(limit), a)
	if err != nil {
		return nil, err
	}
	var list NotificationList
	if err := ParseData(resp, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetUnreadNotificationCount returns the count of unread notifications.
func (c *Client) GetUnreadNotificationCount(ctx context.Context, a *agent.Agent) (int, error) {
	resp, err := c.Get(ctx, "/api/notifications/unread-count", a)
	if err != nil {
		return 0, err
	}
	var count UnreadCount
	if err := ParseData(resp, &count); err != nil {
		return 0, err
	}
	return count.Count, nil
}

// MarkNotificationAsRead marks a notification as read.
func (c *Client) MarkNotificationAsRead(ctx context.Context, a *agent.Agent, id int64) error {
	_, err := c.Put(ctx, "/api/notifications/"+itoa64(id)+"/read", a, nil)
	return err
}

// MarkAllNotificationsAsRead marks all notifications as read.
func (c *Client) MarkAllNotificationsAsRead(ctx context.Context, a *agent.Agent) error {
	_, err := c.Put(ctx, "/api/notifications/read-all", a, nil)
	return err
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

func itoa64(n int64) string {
	return fmt.Sprintf("%d", n)
}
