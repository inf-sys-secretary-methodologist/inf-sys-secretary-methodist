// Package messages contains user-facing strings written by the
// messaging module — primarily system message content stored on
// conversation history (e.g. "Group created", "User joined the
// chat", "User left the chat"). Centralizing here keeps the
// Clean Architecture UI/messaging concern out of usecase logic
// per CLAUDE.md gate ("UI-строки (сообщения пользователю, тексты
// бота) — в `handler/messages`, `llm/responses` и т.п. НЕ в
// usecase"), and makes future i18n a single point of change.
//
// Mirror к internal/modules/auth/interfaces/http/messages/messages.go
// precedent.
package messages

// System message content written by the messaging usecase on
// conversation lifecycle events. The usecase reads these через
// the SystemMessageTexts value type wired в main.go (DI seam) so
// the usecase package itself stays UI-free.
const (
	// SystemGroupCreated is the system message body written into a
	// newly-created group conversation.
	SystemGroupCreated = "Group created"

	// SystemUserJoined is the system message body written when a
	// participant is added to a group conversation.
	SystemUserJoined = "User joined the chat"

	// SystemUserLeft is the system message body written when a
	// participant leaves a group conversation.
	SystemUserLeft = "User left the chat"
)
