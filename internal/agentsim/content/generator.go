package content

import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"

// Generator generates text content for agent actions.
type Generator interface {
	// DocumentTitle generates a document title.
	DocumentTitle(docType, context string) string
	// DocumentContent generates document body text.
	DocumentContent(docType, title, context string) string
	// ChatMessage generates a chat message.
	ChatMessage(from *agent.Agent, to, topic string) string
	// TaskTitle generates a task title.
	TaskTitle(subject string) string
	// TaskDescription generates a task description.
	TaskDescription(title, context string) string
	// EventTitle generates a schedule event title.
	EventTitle(eventType string) string
	// Comment generates a comment on a document/task/report.
	Comment(about string) string
	// ReportTitle generates a report title.
	ReportTitle(reportType string) string
	// ReportDescription generates a report description.
	ReportDescription(title, context string) string
}
