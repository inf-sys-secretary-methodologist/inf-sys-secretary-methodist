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
// RequireRole(system_admin) middleware in production wiring; once
// applied, it gates the destructive write subgroup.
//
// #283 ADR-2 RED stub: the function accepts adminMW but does NOT yet
// apply it — mirroring the pre-fix state of main.go where writes were
// exposed to any authenticated caller. The companion failing tests
// pin the missing gate; the GREEN commit wires adminMW around the
// write subgroup.
func RegisterDepartmentRoutes(
	group *gin.RouterGroup,
	adminMW gin.HandlerFunc,
	departmentHandler *handlers.DepartmentHandler,
) {
	_ = adminMW // RED stub: gate wired in GREEN commit.

	group.GET("", departmentHandler.List)
	group.GET("/:id", departmentHandler.GetByID)
	group.GET("/:id/children", departmentHandler.GetChildren)
	group.POST("", departmentHandler.Create)
	group.PUT("/:id", departmentHandler.Update)
	group.DELETE("/:id", departmentHandler.Delete)
}
