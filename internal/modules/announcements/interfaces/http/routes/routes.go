// Package routes wires the announcements module HTTP endpoints under
// the shared protected group.
//
// v0.163.0 ADR-1 (#303 TIER 0): pre-fix, ALL routes lived flat under
// announcementsGroup with only a JWT gate — including POST/PUT/DELETE
// + publish/unpublish/archive + attachment upload/delete. That let a
// student PUBLISH admin-broadcasts (`POST /api/announcements
// {target_audience:"admins"}`) and tamper with anyone's announcement.
// Split mirrors documents/curriculum/reports v0.133.0 pattern.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/interfaces/http/handlers"
)

// RegisterAnnouncementRoutes mounts the /announcements routes under
// the given protected group. The nonStudentMW handler must be the
// production RequireNonStudent() middleware; it gates the mutation
// subgroup so that students can only READ announcements, not author
// or modify them.
//
// Read endpoints (List / GetByID / GetPublished / GetPinned / GetRecent)
// stay on the parent group — students need to consume announcements.
func RegisterAnnouncementRoutes(
	group *gin.RouterGroup,
	nonStudentMW gin.HandlerFunc,
	handler *handlers.AnnouncementHandler,
) {
	// Permissive read subgroup — any authenticated caller including
	// students. Audience filtering лежит на usecase/repo level (ADR-2).
	group.GET("", handler.List)
	group.GET("/:id", handler.GetByID)
	group.GET("/published", handler.GetPublished)
	group.GET("/pinned", handler.GetPinned)
	group.GET("/recent", handler.GetRecent)

	// Mutation subgroup — non-student only. Closes the TIER 0
	// privilege-escalation gap.
	mut := group.Group("")
	mut.Use(nonStudentMW)
	mut.POST("", handler.Create)
	mut.PUT("/:id", handler.Update)
	mut.DELETE("/:id", handler.Delete)
	mut.POST("/:id/publish", handler.Publish)
	mut.POST("/:id/unpublish", handler.Unpublish)
	mut.POST("/:id/archive", handler.Archive)
	mut.POST("/:id/attachments", handler.UploadAttachment)
	mut.DELETE("/:id/attachments/:attachmentID", handler.DeleteAttachment)
}
