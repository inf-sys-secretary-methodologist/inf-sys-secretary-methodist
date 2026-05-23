package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/interfaces/http/handlers"
)

// RegisterPositionRoutes mounts the /positions routes under the given
// protected group. Identical shape to [RegisterDepartmentRoutes] —
// writes gated by adminMW, reads permissive.
//
// Closes #283 ADR-2 (TIER 0): pre-v0.160.0, all writes were exposed
// to any authenticated caller — same root cause as departments
// (v0.133.0 admin-gate split skipped this group).
func RegisterPositionRoutes(
	group *gin.RouterGroup,
	adminMW gin.HandlerFunc,
	positionHandler *handlers.PositionHandler,
) {
	// Permissive subgroup — read access for cross-module consumers
	// and frontend dropdowns.
	group.GET("", positionHandler.List)
	group.GET("/:id", positionHandler.GetByID)

	// Admin-write subgroup — only system_admin can mutate.
	admin := group.Group("")
	admin.Use(adminMW)
	admin.POST("", positionHandler.Create)
	admin.PUT("/:id", positionHandler.Update)
	admin.DELETE("/:id", positionHandler.Delete)
}
