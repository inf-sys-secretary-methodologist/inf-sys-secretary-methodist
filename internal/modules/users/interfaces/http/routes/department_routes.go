// Package routes wires the users module HTTP endpoints under the
// shared protected group.
//
// Department + position CRUD live here alongside user routes because
// they share the v0.133.0 admin-gate pattern: writes (Create / Update /
// Delete) should live behind the production RequireRole(system_admin)
// middleware; reads (List / GetByID / GetChildren) stay permissive so
// cross-module consumers (curriculum methodist resolver, documents
// author lookup) can reach them.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/interfaces/http/handlers"
)

// RegisterDepartmentRoutes mounts the /departments routes under the
// given protected group. The adminMW handler must be the production
// RequireRole(system_admin) middleware; it gates the destructive
// write subgroup so that only system_admin can mutate the
// organizational structure.
//
// Closes #283 ADR-2 (TIER 0): pre-v0.160.0, all writes were exposed
// to any authenticated caller — including students — because the
// v0.133.0 admin-gate split was applied only to /users, leaving
// /departments and /positions on the permissive parent group.
//
// Read endpoints (List / GetByID / GetChildren) stay on the parent
// group — frontend dropdowns and cross-module resolvers depend on
// the open read surface.
func RegisterDepartmentRoutes(
	group *gin.RouterGroup,
	adminMW gin.HandlerFunc,
	departmentHandler *handlers.DepartmentHandler,
) {
	// Permissive subgroup — any authenticated caller.
	group.GET("", departmentHandler.List)
	group.GET("/:id", departmentHandler.GetByID)
	group.GET("/:id/children", departmentHandler.GetChildren)

	// Admin-write subgroup — only system_admin (adminMW gates the
	// whole subgroup). Closes the TIER 0 privilege-escalation gap.
	admin := group.Group("")
	admin.Use(adminMW)
	admin.POST("", departmentHandler.Create)
	admin.PUT("/:id", departmentHandler.Update)
	admin.DELETE("/:id", departmentHandler.Delete)
}
