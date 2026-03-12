package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/api"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/content"
)

const roleMethodist = "methodist"

func init() {
	Register(DocumentFlowScenario())
}

// DocumentFlowScenario creates the "Document workflow" scenario.
func DocumentFlowScenario() *Scenario {
	return &Scenario{
		Name:        "document-flow",
		Description: "Документооборот — создание, согласование и утверждение документа",
		Steps: []Step{
			{
				Name:  "Секретарь получает типы документов",
				Agent: "Марина Петровна Соколова",
				Delay: 2 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					types, err := c.GetDocumentTypes(ctx, a)
					if err != nil {
						return fmt.Errorf("get document types: %w", err)
					}
					// Store first type ID for later use
					if len(types) > 0 {
						state.SetExtra("doc_type_id", types[0].ID)
					}
					return nil
				},
			},
			{
				Name:  "Секретарь создаёт приказ",
				Agent: "Марина Петровна Соколова",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					docTypeID := int64(1)
					if v, ok := state.GetExtra("doc_type_id"); ok {
						if id, ok := v.(int64); ok {
							docTypeID = id
						}
					}

					title := gen.DocumentTitle("order", "начало семестра")
					body := gen.DocumentContent("order", title, "начало семестра")

					doc, err := c.CreateDocument(ctx, a, api.CreateDocumentRequest{
						Title:          title,
						DocumentTypeID: docTypeID,
						Content:        body,
						Subject:        "Организация учебного процесса",
						Importance:     "high",
					})
					if err != nil {
						return fmt.Errorf("create document: %w", err)
					}

					state.SetDoc("order_1", doc.ID)
					return nil
				},
			},
			{
				Name:  "Секретарь создаёт служебную записку",
				Agent: "Марина Петровна Соколова",
				Delay: 45 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					docTypeID := int64(1)
					if v, ok := state.GetExtra("doc_type_id"); ok {
						if id, ok := v.(int64); ok {
							docTypeID = id
						}
					}

					title := gen.DocumentTitle("memo", "текущие вопросы")
					body := gen.DocumentContent("memo", title, "текущие вопросы")

					doc, err := c.CreateDocument(ctx, a, api.CreateDocumentRequest{
						Title:          title,
						DocumentTypeID: docTypeID,
						Content:        body,
						Subject:        "Текущие организационные вопросы",
						Importance:     "normal",
					})
					if err != nil {
						return fmt.Errorf("create memo: %w", err)
					}

					state.SetDoc("memo_1", doc.ID)
					return nil
				},
			},
			{
				Name:  "Секретарь делится приказом с методистом",
				Agent: "Марина Петровна Соколова",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					docID, ok := state.GetDoc("order_1")
					if !ok {
						return fmt.Errorf("order_1 not found in state")
					}

					methodistID := findAgentByRole(state, roleMethodist)
					if methodistID == 0 {
						// Try getting users list
						users, err := c.ListUsers(ctx, a)
						if err != nil {
							return fmt.Errorf("list users: %w", err)
						}
						for _, u := range users.Users {
							if u.Role == roleMethodist {
								methodistID = u.ID
								state.SetExtra("user_methodist", u.ID)
								break
							}
						}
					}

					if methodistID == 0 {
						return fmt.Errorf("no methodist found")
					}

					return c.ShareDocument(ctx, a, docID, methodistID, "read")
				},
			},
			{
				Name:  "Методист просматривает документ",
				Agent: "Алексей Николаевич Козлов",
				Delay: 60 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					docID, ok := state.GetDoc("order_1")
					if !ok {
						return fmt.Errorf("order_1 not found in state")
					}

					doc, err := c.GetDocument(ctx, a, docID)
					if err != nil {
						return fmt.Errorf("get document: %w", err)
					}
					_ = doc
					return nil
				},
			},
			{
				Name:  "Методист проверяет список документов",
				Agent: "Алексей Николаевич Козлов",
				Delay: 30 * time.Second,
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
				Name:  "Админ просматривает документ",
				Agent: "Елена Сергеевна Иванова",
				Delay: 45 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					docID, ok := state.GetDoc("order_1")
					if !ok {
						return fmt.Errorf("order_1 not found in state")
					}

					doc, err := c.GetDocument(ctx, a, docID)
					if err != nil {
						return fmt.Errorf("get document: %w", err)
					}
					_ = doc
					return nil
				},
			},
			{
				Name:  "Секретарь уведомляет в чате об утверждении",
				Agent: "Марина Петровна Соколова",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					// Get or create a conversation
					convID, ok := state.GetConversation("document_flow_chat")
					if !ok {
						methodistID := findAgentByRole(state, roleMethodist)
						if methodistID == 0 {
							return fmt.Errorf("methodist not found")
						}
						conv, err := c.CreateDirectConversation(ctx, a, methodistID)
						if err != nil {
							return fmt.Errorf("create conversation: %w", err)
						}
						convID = conv.ID
						state.SetConversation("document_flow_chat", convID)
					}

					_, err := c.SendMessage(ctx, a, convID, "Приказ подготовлен и направлен на ознакомление. Прошу проверить.")
					return err
				},
			},
		},
	}
}
