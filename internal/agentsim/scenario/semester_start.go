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
	Register(SemesterStartScenario())
}

// SemesterStartScenario creates the "Semester start" scenario.
func SemesterStartScenario() *Scenario {
	return &Scenario{
		Name:        "semester-start",
		Description: "Начало семестра — расписание, документы, задания",
		Steps: []Step{
			{
				Name:  "Методист создаёт события расписания (пн-пт)",
				Agent: "Алексей Николаевич Козлов",
				Delay: 5 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					now := time.Now()
					weekday := int(now.Weekday())
					if weekday == 0 {
						weekday = 7
					}
					// Start from Monday of this week
					monday := now.AddDate(0, 0, -weekday+1)

					subjects := []string{
						"Высшая математика", "Русский язык и литература",
						"Общая физика", "Информатика", "Философия",
					}

					for day := 0; day < 5; day++ {
						date := monday.AddDate(0, 0, day)
						subj := subjects[day%len(subjects)]
						startTime := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, date.Location())
						endTime := startTime.Add(90 * time.Minute)

						title := gen.EventTitle("task")
						if title == "" {
							title = subj
						}

						event, err := c.CreateEvent(ctx, a, api.CreateEventRequest{
							Title:     title,
							EventType: "task",
							StartTime: startTime.Format(time.RFC3339),
							EndTime:   endTime.Format(time.RFC3339),
							Location:  fmt.Sprintf("Аудитория %d", 101+day),
						})
						if err != nil {
							continue
						}
						state.SetEvent(fmt.Sprintf("class_day_%d", day+1), event.ID)
					}
					return nil
				},
			},
			{
				Name:  "Секретарь создаёт служебные записки",
				Agent: "Марина Петровна Соколова",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					docTypeID := int64(1)
					if v, ok := state.GetExtra("doc_type_id"); ok {
						if id, ok := v.(int64); ok {
							docTypeID = id
						}
					}

					for i := 0; i < 3; i++ {
						title := gen.DocumentTitle("memo", "начало семестра")
						body := gen.DocumentContent("memo", title, "начало семестра")

						doc, err := c.CreateDocument(ctx, a, api.CreateDocumentRequest{
							Title:          title,
							DocumentTypeID: docTypeID,
							Content:        body,
							Subject:        "Организация начала семестра",
							Importance:     "normal",
						})
						if err != nil {
							continue
						}
						state.SetDoc(fmt.Sprintf("semester_memo_%d", i+1), doc.ID)

						// Small delay between document creation
						select {
						case <-ctx.Done():
							return ctx.Err()
						case <-time.After(5 * time.Second):
						}
					}
					return nil
				},
			},
			{
				Name:  "Преподаватель математики создаёт задание",
				Agent: "Дмитрий Иванович Волков",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					title := gen.TaskTitle("Контрольная работа №1 по высшей математике")
					desc := gen.TaskDescription(title, "начало семестра, первое задание")

					task, err := c.CreateTask(ctx, a, api.CreateTaskRequest{
						Title:       title,
						Description: desc,
						Priority:    "high",
						DueDate:     time.Now().AddDate(0, 0, 14).Format("2006-01-02"),
					})
					if err != nil {
						return fmt.Errorf("create task: %w", err)
					}
					state.SetTask("math_task_1", task.ID)
					return nil
				},
			},
			{
				Name:  "Преподаватель литературы создаёт задание",
				Agent: "Ольга Витальевна Михайлова",
				Delay: 45 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					title := gen.TaskTitle("Эссе по русской литературе XIX века")
					desc := gen.TaskDescription(title, "начало семестра")

					task, err := c.CreateTask(ctx, a, api.CreateTaskRequest{
						Title:       title,
						Description: desc,
						Priority:    "normal",
						DueDate:     time.Now().AddDate(0, 0, 21).Format("2006-01-02"),
					})
					if err != nil {
						return fmt.Errorf("create task: %w", err)
					}
					state.SetTask("literature_task_1", task.ID)
					return nil
				},
			},
			{
				Name:  "Студент проверяет уведомления",
				Agent: "Иван Смирнов",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					notifications, err := c.ListNotifications(ctx, a, 20)
					if err != nil {
						return fmt.Errorf("list notifications: %w", err)
					}
					_ = notifications
					return nil
				},
			},
			{
				Name:  "Студентка проверяет расписание",
				Agent: "Анна Кузнецова",
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
				Name:  "Преподаватель пишет в группу преподавателей",
				Agent: "Дмитрий Иванович Волков",
				Delay: 45 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					// List existing conversations or create group
					convs, err := c.ListConversations(ctx, a)
					if err != nil {
						return fmt.Errorf("list conversations: %w", err)
					}

					var convID int64
					for _, conv := range convs.Conversations {
						if conv.Type == "group" {
							convID = conv.ID
							break
						}
					}

					if convID == 0 {
						// No group conversation yet, send to existing or skip
						if len(convs.Conversations) > 0 {
							convID = convs.Conversations[0].ID
						} else {
							return nil // No conversations available
						}
					}

					msg := gen.ChatMessage(a, "преподаватели", "начало семестра")
					_, err = c.SendMessage(ctx, a, convID, msg)
					return err
				},
			},
			{
				Name:  "Секретарь создаёт протокол совещания",
				Agent: "Марина Петровна Соколова",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					docTypeID := int64(1)
					if v, ok := state.GetExtra("doc_type_id"); ok {
						if id, ok := v.(int64); ok {
							docTypeID = id
						}
					}

					title := gen.DocumentTitle("protocol", "начало семестра")
					body := gen.DocumentContent("protocol", title, "начало семестра")

					doc, err := c.CreateDocument(ctx, a, api.CreateDocumentRequest{
						Title:          title,
						DocumentTypeID: docTypeID,
						Content:        body,
						Subject:        "Протокол совещания по началу семестра",
						Importance:     "high",
					})
					if err != nil {
						return fmt.Errorf("create protocol: %w", err)
					}
					state.SetDoc("protocol_semester", doc.ID)
					return nil
				},
			},
		},
	}
}
