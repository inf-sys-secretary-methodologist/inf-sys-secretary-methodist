package scenario

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/api"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/content"
)

func init() {
	Register(WorkdayScenario())
}

// WorkdayScenario creates the "Workday" scenario — a mix of everyday activities.
func WorkdayScenario() *Scenario {
	return &Scenario{
		Name:        "workday",
		Description: "Рабочий день — комбинация действий всех ролей",
		Steps: []Step{
			// Morning check-ins
			{
				Name:  "Секретарь: утренняя проверка",
				Agent: "Марина Петровна Соколова",
				Delay: 5 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					_, _ = c.GetUnreadNotificationCount(ctx, a)
					_, _ = c.ListConversations(ctx, a)
					return nil
				},
			},
			{
				Name:  "Методист: утренняя проверка",
				Agent: "Алексей Николаевич Козлов",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					_, _ = c.ListDocuments(ctx, a, "page_size=5")
					_, _ = c.ListTasks(ctx, a, "limit=5")
					return nil
				},
			},
			// Document work
			{
				Name:  "Секретарь создаёт документ",
				Agent: "Марина Петровна Соколова",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					docTypes := []string{"memo", "order", "protocol"}
					docType := docTypes[rand.Intn(len(docTypes))]

					title := gen.DocumentTitle(docType, "рабочий день")
					body := gen.DocumentContent(docType, title, "рабочий день")

					types, _ := c.GetDocumentTypes(ctx, a)
					typeID := int64(1)
					if len(types) > 0 {
						typeID = types[rand.Intn(len(types))].ID
					}

					doc, err := c.CreateDocument(ctx, a, api.CreateDocumentRequest{
						Title:          title,
						DocumentTypeID: typeID,
						Content:        body,
						Importance:     "normal",
					})
					if err != nil {
						return err
					}
					state.SetDoc("workday_doc", doc.ID)
					return nil
				},
			},
			// Teacher creates a task
			{
				Name:  "Преподаватель создаёт задание",
				Agent: "Сергей Павлович Новиков",
				Delay: 90 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					title := gen.TaskTitle("")
					desc := gen.TaskDescription(title, "рабочий день")

					task, err := c.CreateTask(ctx, a, api.CreateTaskRequest{
						Title:       title,
						Description: desc,
						Priority:    "normal",
						DueDate:     time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
					})
					if err != nil {
						return err
					}
					state.SetTask("workday_task", task.ID)
					return nil
				},
			},
			// Messaging between colleagues
			{
				Name:  "Секретарь пишет админу",
				Agent: "Марина Петровна Соколова",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					adminID := findAgentByRole(state, "system_admin")
					if adminID == 0 {
						users, err := c.ListUsers(ctx, a)
						if err != nil {
							return err
						}
						for _, u := range users.Users {
							if u.Role == "system_admin" {
								adminID = u.ID
								state.SetExtra("user_system_admin", u.ID)
								break
							}
						}
					}
					if adminID == 0 {
						return nil
					}

					conv, err := c.CreateDirectConversation(ctx, a, adminID)
					if err != nil {
						return err
					}
					state.SetConversation("secretary_admin", conv.ID)

					msg := gen.ChatMessage(a, "админ", "рабочие вопросы")
					_, err = c.SendMessage(ctx, a, conv.ID, msg)
					return err
				},
			},
			{
				Name:  "Админ отвечает секретарю",
				Agent: "Елена Сергеевна Иванова",
				Delay: 45 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					convID, ok := state.GetConversation("secretary_admin")
					if !ok {
						convs, err := c.ListConversations(ctx, a)
						if err != nil || len(convs.Conversations) == 0 {
							return nil
						}
						convID = convs.Conversations[0].ID
					}

					msg := gen.ChatMessage(a, "секретарь", "рабочие вопросы")
					_, err := c.SendMessage(ctx, a, convID, msg)
					return err
				},
			},
			// Student activity
			{
				Name:  "Студент проверяет задания и расписание",
				Agent: "Анна Кузнецова",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					_, _ = c.ListTasks(ctx, a, fmt.Sprintf("assignee_id=%d&limit=5", a.UserID))
					_, _ = c.GetUpcomingEvents(ctx, a)
					_, _ = c.GetUnreadNotificationCount(ctx, a)
					return nil
				},
			},
			// Methodist reviews documents
			{
				Name:  "Методист просматривает документы",
				Agent: "Алексей Николаевич Козлов",
				Delay: 90 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					docs, err := c.ListDocuments(ctx, a, "page_size=5")
					if err != nil {
						return err
					}
					// Read first few documents
					for i, doc := range docs.Documents {
						if i >= 3 {
							break
						}
						_, _ = c.GetDocument(ctx, a, doc.ID)
					}
					return nil
				},
			},
			// Teacher checks messages
			{
				Name:  "Преподаватель проверяет сообщения",
				Agent: "Ольга Витальевна Михайлова",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					convs, err := c.ListConversations(ctx, a)
					if err != nil {
						return err
					}
					for _, conv := range convs.Conversations {
						if conv.UnreadCount > 0 {
							msgs, err := c.GetMessages(ctx, a, conv.ID, 5)
							if err != nil {
								continue
							}
							if len(msgs.Messages) > 0 {
								_ = c.MarkConversationAsRead(ctx, a, conv.ID, msgs.Messages[0].ID)
							}
						}
					}
					return nil
				},
			},
			// End of day notifications check
			{
				Name:  "Секретарь: вечерняя проверка уведомлений",
				Agent: "Марина Петровна Соколова",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					notifications, err := c.ListNotifications(ctx, a, 20)
					if err != nil {
						return err
					}
					// Mark all as read
					if notifications.UnreadCount > 0 {
						_ = c.MarkAllNotificationsAsRead(ctx, a)
					}
					return nil
				},
			},
		},
	}
}
