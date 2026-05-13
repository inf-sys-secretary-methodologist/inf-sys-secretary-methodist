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
	// RED state: every endpoint mounted on the parent group without
	// the admin gate. Tests in routes_test.go assert that the four
	// non-admin roles must receive 403 on write endpoints — those
	// assertions FAIL here and PASS once GREEN splits the writes
	// under adminMW.
	group.GET("", userHandler.List)
	group.GET("/:id", userHandler.GetByID)
	group.PUT("/:id/profile", userHandler.UpdateProfile)
	group.PUT("/:id/role", userHandler.UpdateRole)
	group.PUT("/:id/status", userHandler.UpdateStatus)
	group.DELETE("/:id", userHandler.Delete)
	group.POST("/bulk/department", userHandler.BulkUpdateDepartment)
	group.POST("/bulk/position", userHandler.BulkUpdatePosition)
	group.GET("/by-department/:id", userHandler.GetByDepartment)
	group.GET("/by-position/:id", userHandler.GetByPosition)
	group.POST("/:id/avatar", avatarHandler.Upload)
	group.DELETE("/:id/avatar", avatarHandler.Delete)
	group.GET("/:id/avatar", avatarHandler.GetAvatarURL)

	// adminMW is the RequireRole(system_admin) gate. In the RED
	// commit it stays unused — declared in the signature so callers
	// can already pass it without a follow-up signature change. The
	// GREEN commit applies it to the write subgroup.
	_ = adminMW
}
