package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/api"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/content"
)

func init() {
	Register(MorningScenario())
}

// MorningScenario creates the "Morning at the university" scenario.
func MorningScenario() *Scenario {
	return &Scenario{
		Name:        "morning",
		Description: "Утро в вузе — сотрудники начинают рабочий день",
		Steps: []Step{
			{
				Name:  "Секретарь проверяет уведомления",
				Agent: "Марина Петровна Соколова",
				Delay: 2 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					count, err := c.GetUnreadNotificationCount(ctx, a)
					if err != nil {
						return fmt.Errorf("get unread count: %w", err)
					}
					_ = count
					return nil
				},
			},
			{
				Name:  "Секретарь читает сообщения",
				Agent: "Марина Петровна Соколова",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					convs, err := c.ListConversations(ctx, a)
					if err != nil {
						return fmt.Errorf("list conversations: %w", err)
					}
					// Read messages from first conversation with unread messages
					for _, conv := range convs.Conversations {
						if conv.UnreadCount > 0 {
							msgs, err := c.GetMessages(ctx, a, conv.ID, 10)
							if err != nil {
								continue
							}
							if len(msgs.Messages) > 0 {
								_ = c.MarkConversationAsRead(ctx, a, conv.ID, msgs.Messages[0].ID)
							}
							break
						}
					}
					return nil
				},
			},
			{
				Name:  "Методист логинится и смотрит документы",
				Agent: "Алексей Николаевич Козлов",
				Delay: 45 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					docs, err := c.ListDocuments(ctx, a, "page_size=10")
					if err != nil {
						return fmt.Errorf("list documents: %w", err)
					}
					_ = docs
					return nil
				},
			},
			{
				Name:  "Преподаватель проверяет расписание",
				Agent: "Дмитрий Иванович Волков",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					events, err := c.GetUpcomingEvents(ctx, a)
					if err != nil {
						return fmt.Errorf("get upcoming events: %w", err)
					}
					_ = events
					return nil
				},
			},
			{
				Name:  "Второй преподаватель проверяет расписание",
				Agent: "Ольга Витальевна Михайлова",
				Delay: 20 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					events, err := c.GetUpcomingEvents(ctx, a)
					if err != nil {
						return fmt.Errorf("get upcoming events: %w", err)
					}
					_ = events
					return nil
				},
			},
			{
				Name:  "Студент проверяет задания",
				Agent: "Иван Смирнов",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					tasks, err := c.ListTasks(ctx, a, fmt.Sprintf("assignee_id=%d&limit=10", a.UserID))
					if err != nil {
						return fmt.Errorf("list tasks: %w", err)
					}
					_ = tasks
					return nil
				},
			},
			{
				Name:  "Секретарь пишет методисту",
				Agent: "Марина Петровна Соколова",
				Delay: 45 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					// Find methodist agent to get their user ID
					methodist := findAgentByRole(state, "methodist")
					if methodist == 0 {
						// Store methodist ID for later use by looking it up
						return fmt.Errorf("methodist user ID not available")
					}

					conv, err := c.CreateDirectConversation(ctx, a, methodist)
					if err != nil {
						return fmt.Errorf("create conversation: %w", err)
					}
					state.SetConversation("secretary_methodist", conv.ID)

					msg := gen.ChatMessage(a, "методист", "утренние дела")
					_, err = c.SendMessage(ctx, a, conv.ID, msg)
					return err
				},
			},
			{
				Name:  "Методист отвечает секретарю",
				Agent: "Алексей Николаевич Козлов",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					convID, ok := state.GetConversation("secretary_methodist")
					if !ok {
						// Try listing conversations instead
						convs, err := c.ListConversations(ctx, a)
						if err != nil || len(convs.Conversations) == 0 {
							return fmt.Errorf("no conversations found")
						}
						convID = convs.Conversations[0].ID
					}

					msg := gen.ChatMessage(a, "секретарь", "утренние дела")
					_, err := c.SendMessage(ctx, a, convID, msg)
					return err
				},
			},
		},
	}
}

// findAgentByRole is a helper that retrieves a stored user ID from state.
func findAgentByRole(state *SharedState, role string) int64 {
	val, ok := state.GetExtra("user_" + role)
	if !ok {
		return 0
	}
	id, ok := val.(int64)
	if !ok {
		return 0
	}
	return id
}
