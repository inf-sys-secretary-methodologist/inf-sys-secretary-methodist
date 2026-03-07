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
	Register(MeetingScenario())
}

// MeetingScenario creates the "Meeting" scenario.
func MeetingScenario() *Scenario {
	return &Scenario{
		Name:        "meeting",
		Description: "Совещание — создание, проведение, протокол",
		Steps: []Step{
			{
				Name:  "Секретарь находит участников совещания",
				Agent: "Марина Петровна Соколова",
				Delay: 5 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					users, err := c.ListUsers(ctx, a)
					if err != nil {
						return fmt.Errorf("list users: %w", err)
					}
					// Store participant IDs for later
					var participantIDs []int64
					for _, u := range users.Users {
						if u.Role == "methodist" || u.Role == "teacher" || u.Role == "system_admin" {
							participantIDs = append(participantIDs, u.ID)
							state.SetExtra("user_"+u.Role, u.ID)
						}
						if len(participantIDs) >= 5 {
							break
						}
					}
					state.SetExtra("meeting_participants", participantIDs)
					return nil
				},
			},
			{
				Name:  "Секретарь создаёт событие-совещание",
				Agent: "Марина Петровна Соколова",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					title := gen.EventTitle("meeting")

					startTime := time.Now().Add(2 * time.Hour)
					endTime := startTime.Add(90 * time.Minute)

					var participantIDs []int64
					if v, ok := state.GetExtra("meeting_participants"); ok {
						if ids, ok := v.([]int64); ok {
							participantIDs = ids
						}
					}

					event, err := c.CreateEvent(ctx, a, api.CreateEventRequest{
						Title:          title,
						Description:    "Совещание по текущим вопросам учебного процесса",
						EventType:      "meeting",
						StartTime:      startTime.Format(time.RFC3339),
						EndTime:        endTime.Format(time.RFC3339),
						Location:       "Конференц-зал, каб. 301",
						ParticipantIDs: participantIDs,
						Priority:       "high",
					})
					if err != nil {
						return fmt.Errorf("create meeting event: %w", err)
					}
					state.SetEvent("meeting", event.ID)
					return nil
				},
			},
			{
				Name:  "Методист принимает приглашение",
				Agent: "Алексей Николаевич Козлов",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					eventID, ok := state.GetEvent("meeting")
					if !ok {
						return fmt.Errorf("meeting event not found")
					}
					return c.RespondToEvent(ctx, a, eventID, "accepted")
				},
			},
			{
				Name:  "Преподаватель принимает приглашение",
				Agent: "Дмитрий Иванович Волков",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					eventID, ok := state.GetEvent("meeting")
					if !ok {
						return fmt.Errorf("meeting event not found")
					}
					return c.RespondToEvent(ctx, a, eventID, "accepted")
				},
			},
			{
				Name:  "Админ принимает приглашение",
				Agent: "Елена Сергеевна Иванова",
				Delay: 20 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					eventID, ok := state.GetEvent("meeting")
					if !ok {
						return fmt.Errorf("meeting event not found")
					}
					return c.RespondToEvent(ctx, a, eventID, "accepted")
				},
			},
			{
				Name:  "Секретарь создаёт протокол совещания",
				Agent: "Марина Петровна Соколова",
				Delay: 90 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					types, _ := c.GetDocumentTypes(ctx, a)
					typeID := int64(1)
					if len(types) > 0 {
						typeID = types[0].ID
					}

					title := gen.DocumentTitle("protocol", "совещание")
					body := gen.DocumentContent("protocol", title, "совещание по текущим вопросам")

					doc, err := c.CreateDocument(ctx, a, api.CreateDocumentRequest{
						Title:          title,
						DocumentTypeID: typeID,
						Content:        body,
						Subject:        "Протокол совещания",
						Importance:     "high",
					})
					if err != nil {
						return fmt.Errorf("create protocol: %w", err)
					}
					state.SetDoc("meeting_protocol", doc.ID)
					return nil
				},
			},
			{
				Name:  "Секретарь рассылает ссылку на протокол",
				Agent: "Марина Петровна Соколова",
				Delay: 45 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					methodistID := findAgentByRole(state, "methodist")
					if methodistID == 0 {
						return nil
					}

					conv, err := c.CreateDirectConversation(ctx, a, methodistID)
					if err != nil {
						return err
					}

					msg := "Протокол совещания готов и загружен в систему. Прошу ознакомиться и при необходимости дополнить."
					_, err = c.SendMessage(ctx, a, conv.ID, msg)
					return err
				},
			},
		},
	}
}
