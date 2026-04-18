package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRoles returns a middleware that enforces role-based access control.
// It must be used AFTER TenantMiddleware since it reads CtxUserRole.
func RequireRoles(allowed ...string) gin.HandlerFunc {
	set := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		set[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, _ := c.Get(CtxUserRole)
		roleStr, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role not determined"})
			return
		}

		if _, permitted := set[roleStr]; !permitted {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}
