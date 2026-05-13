// Package routes wires the users module HTTP endpoints under the
// shared protected group. The function accepts an admin middleware so
// that destructive endpoints (role/status/delete/bulk/avatar mutations)
// are gated by RequireRole(system_admin) in production. Read-only
// endpoints stay permissive — cross-module consumers (documents author
// lookup, curriculum methodist resolver) need them.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/interfaces/http/handlers"
)

// RegisterUserRoutes mounts the /users routes under the given protected
// group. The adminMW handler must be the production
// RequireRole(system_admin) middleware in production wiring; it gates
// the destructive write subgroup so that only system_admin can mutate
// other users' role/status/profile or invoke bulk operations.
//
// Read-only endpoints (List, GetByID, GetByDepartment, GetByPosition,
// avatar GET) stay on the parent group — any authenticated caller may
// reach them, mirroring the cross-module read needs of documents and
// curriculum.
func RegisterUserRoutes(
	group *gin.RouterGroup,
	adminMW gin.HandlerFunc,
	userHandler *handlers.UserHandler,
	avatarHandler *handlers.AvatarHandler,
) {
	// Read-only subgroup — any authenticated caller. Cross-module
	// consumers (documents author lookup, curriculum methodist
	// resolver) depend on this surface and must not be admin-gated.
	group.GET("", userHandler.List)
	group.GET("/:id", userHandler.GetByID)
	group.GET("/by-department/:id", userHandler.GetByDepartment)
	group.GET("/by-position/:id", userHandler.GetByPosition)
	group.GET("/:id/avatar", avatarHandler.GetAvatarURL)

	// Admin-write subgroup — only system_admin (adminMW gates the
	// whole subgroup). Closes the TIER 0 privilege-escalation gap
	// where any authenticated user could DELETE /:id, PUT /:id/role,
	// PUT /:id/status, PUT /:id/profile, or invoke /bulk/*.
	admin := group.Group("")
	admin.Use(adminMW)
	admin.PUT("/:id/profile", userHandler.UpdateProfile)
	admin.PUT("/:id/role", userHandler.UpdateRole)
	admin.PUT("/:id/status", userHandler.UpdateStatus)
	admin.DELETE("/:id", userHandler.Delete)
	admin.POST("/bulk/department", userHandler.BulkUpdateDepartment)
	admin.POST("/bulk/position", userHandler.BulkUpdatePosition)
	admin.POST("/:id/avatar", avatarHandler.Upload)
	admin.DELETE("/:id/avatar", avatarHandler.Delete)
}
