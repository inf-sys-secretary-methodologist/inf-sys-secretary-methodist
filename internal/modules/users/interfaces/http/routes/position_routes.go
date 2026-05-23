package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/interfaces/http/handlers"
)

// RegisterPositionRoutes mounts the /positions routes under the given
// protected group. Identical shape to [RegisterDepartmentRoutes] —
// writes should be gated by adminMW, reads stay permissive.
//
// #283 ADR-2 RED stub: gate not yet applied; companion failing tests
// pin the missing gate; GREEN commit wires adminMW.
func RegisterPositionRoutes(
	group *gin.RouterGroup,
	adminMW gin.HandlerFunc,
	positionHandler *handlers.PositionHandler,
) {
	_ = adminMW // RED stub: gate wired in GREEN commit.

	group.GET("", positionHandler.List)
	group.GET("/:id", positionHandler.GetByID)
	group.POST("", positionHandler.Create)
	group.PUT("/:id", positionHandler.Update)
	group.DELETE("/:id", positionHandler.Delete)
}
