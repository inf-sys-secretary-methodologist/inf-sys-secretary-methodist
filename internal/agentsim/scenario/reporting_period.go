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
	Register(ReportingPeriodScenario())
}

// ReportingPeriodScenario creates the "Reporting period" scenario.
func ReportingPeriodScenario() *Scenario {
	return &Scenario{
		Name:        "reporting-period",
		Description: "Отчётный период — создание, генерация, рецензия, публикация",
		Steps: []Step{
			{
				Name:  "Методист получает типы отчётов",
				Agent: "Алексей Николаевич Козлов",
				Delay: 5 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					types, err := c.GetReportTypes(ctx, a)
					if err != nil {
						return fmt.Errorf("get report types: %w", err)
					}
					if len(types) > 0 {
						state.SetExtra("report_type_id", types[0].ID)
					}
					return nil
				},
			},
			{
				Name:  "Методист создаёт отчёт",
				Agent: "Алексей Николаевич Козлов",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					reportTypeID := int64(1)
					if v, ok := state.GetExtra("report_type_id"); ok {
						if id, ok := v.(int64); ok {
							reportTypeID = id
						}
					}

					now := time.Now()
					periodStart := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
					periodEnd := periodStart.AddDate(0, 1, -1)

					title := gen.ReportTitle("monthly")
					desc := gen.ReportDescription(title, "ежемесячный отчёт")

					report, err := c.CreateReport(ctx, a, api.CreateReportRequest{
						ReportTypeID: reportTypeID,
						Title:        title,
						Description:  desc,
						PeriodStart:  periodStart.Format("2006-01-02"),
						PeriodEnd:    periodEnd.Format("2006-01-02"),
					})
					if err != nil {
						return fmt.Errorf("create report: %w", err)
					}
					state.SetReport("monthly_report", report.ID)
					return nil
				},
			},
			{
				Name:  "Методист генерирует данные отчёта",
				Agent: "Алексей Николаевич Козлов",
				Delay: 45 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					reportID, ok := state.GetReport("monthly_report")
					if !ok {
						return fmt.Errorf("monthly_report not found")
					}
					return c.GenerateReport(ctx, a, reportID)
				},
			},
			{
				Name:  "Методист отправляет на рецензию",
				Agent: "Алексей Николаевич Козлов",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					reportID, ok := state.GetReport("monthly_report")
					if !ok {
						return fmt.Errorf("monthly_report not found")
					}
					return c.SubmitReport(ctx, a, reportID)
				},
			},
			{
				Name:  "Админ рецензирует отчёт",
				Agent: "Елена Сергеевна Иванова",
				Delay: 90 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					reportID, ok := state.GetReport("monthly_report")
					if !ok {
						return fmt.Errorf("monthly_report not found")
					}

					comment := gen.Comment("рецензия отчёта")
					return c.ReviewReport(ctx, a, reportID, "approve", comment)
				},
			},
			{
				Name:  "Методист публикует отчёт",
				Agent: "Алексей Николаевич Козлов",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					reportID, ok := state.GetReport("monthly_report")
					if !ok {
						return fmt.Errorf("monthly_report not found")
					}
					return c.PublishReport(ctx, a, reportID, true)
				},
			},
			{
				Name:  "Методист уведомляет в чате",
				Agent: "Алексей Николаевич Козлов",
				Delay: 30 * time.Second,
				Action: func(ctx context.Context, a *agent.Agent, c *api.Client, state *SharedState, gen content.Generator) error {
					secretaryID := findAgentByRole(state, "academic_secretary")
					if secretaryID == 0 {
						users, err := c.ListUsers(ctx, a)
						if err != nil {
							return err
						}
						for _, u := range users.Users {
							if u.Role == "academic_secretary" {
								secretaryID = u.ID
								state.SetExtra("user_academic_secretary", u.ID)
								break
							}
						}
					}
					if secretaryID == 0 {
						return nil
					}

					conv, err := c.CreateDirectConversation(ctx, a, secretaryID)
					if err != nil {
						return err
					}

					_, err = c.SendMessage(ctx, a, conv.ID, "Ежемесячный отчёт утверждён и опубликован. Доступен для ознакомления в разделе отчётов.")
					return err
				},
			},
		},
	}
}
