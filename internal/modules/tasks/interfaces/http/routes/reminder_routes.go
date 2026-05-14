// Package routes wires the task reminder HTTP endpoints under the
// caller-provided protected group. Greenfield в v0.138.0 — mirror к
// v0.137.1 branding.RegisterBrandingRoutes registrar shape.
//
// The registrar does NOT take an adminMW parameter — all reminder
// endpoints are user-self-scoped (caller's user_id = JWT subject),
// any authenticated role can manage their own reminders. Per-user
// privacy comes from each use case taking ActorUserID, not from a
// role gate.
package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/interfaces/http/handlers"
)

// RegisterTaskReminderRoutes mounts /tasks/:id/reminders + nested
// /tasks/:id/reminders/:reminderID under the supplied protected
// group. The caller is responsible for applying the JWT auth
// middleware to the group before calling this function (mirror к
// branding registrar contract).
//
// Routes:
//   - POST   /tasks/:id/reminders                  → handler.Create
//   - GET    /tasks/:id/reminders                  → handler.List
//   - DELETE /tasks/:id/reminders/:reminderID      → handler.Delete
//   - OPTIONS /tasks/:id/reminders                 → 204 CORS
//   - OPTIONS /tasks/:id/reminders/:reminderID     → 204 CORS
func RegisterTaskReminderRoutes(
	protectedGroup *gin.RouterGroup,
	handler *handlers.TaskReminderHandler,
) {
	protectedGroup.POST("/tasks/:id/reminders", handler.Create)
	protectedGroup.GET("/tasks/:id/reminders", handler.List)
	protectedGroup.DELETE("/tasks/:id/reminders/:reminderID", handler.Delete)
	protectedGroup.OPTIONS("/tasks/:id/reminders", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	protectedGroup.OPTIONS("/tasks/:id/reminders/:reminderID", func(c *gin.Context) { c.Status(http.StatusNoContent) })
}
