package handlers_test

import "github.com/gin-gonic/gin"

// withAuth installs production-shaped middleware on the engine —
// c.Set("user_id", uid) + c.Set("role", role). Mirrors the keys
// production middleware writes (internal/modules/auth/interfaces/http/
// middleware/auth_middleware.go: c.Set("user_id", ...) + c.Set("role",
// ...)).
//
// Pinning to the same keys prevents the v0.126.0 wrong-key bug class
// from re-emerging — handler reads c.Get("role"); if a test helper
// wrote "user_role" instead, the integration would silently 401 in
// production while tests pretended to pass.
//
// Pass uid=0 to omit the user_id key (simulating an upstream gap).
// Pass role="" to omit the role key (simulating a stripped-context
// auth flow). Both omissions cause the handler's authContext to
// return ok=false → 401 response.
//
// Shared by section_handler_test.go и discipline_item_handler_test.go;
// other curriculum_*_handler_test.go files inline the same middleware
// (pre-existing pattern, не migrate в same release).
func withAuth(uid int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if uid != 0 {
			c.Set("user_id", uid)
		}
		if role != "" {
			c.Set("role", role)
		}
		c.Next()
	}
}
