package middleware

import (
	"net/http"

	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

type PermissionChecker interface {
	HasPermission(roleName, permission string) (bool, error)
}

// AdminOnly gates routes on the literal JWT role claim "admin".
// This codebase uses RequirePermission for admin APIs so RBAC stays in the database.
// Keep AdminOnly for forks or internal routes that intentionally bypass permission rows.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")

		if !exists || role != "admin" {
			httpx.Error(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole gates routes on a single role name from the JWT.
// Prefer RequirePermission for admin surfaces; RequireRole is useful for coarse checks
// (e.g. a dedicated operator role) when permission data is not needed.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")

		if !exists || userRole != role {
			httpx.Error(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequirePermission(checker PermissionChecker, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		roleName, ok := roleValue.(string)
		if !exists || !ok || roleName == "" {
			httpx.Error(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		allowed, err := checker.HasPermission(roleName, permission)
		if err != nil {
			httpx.Error(c, http.StatusInternalServerError, "Failed to check permissions")
			c.Abort()
			return
		}

		if !allowed {
			httpx.Error(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOwnerOrPermission passes if the authenticated user owns the resource
// (owner_user_id matches) OR holds the specified admin permission.
func RequireOwnerOrPermission(checker PermissionChecker, permission string, getOwnerID func(c *gin.Context) (uint, bool)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		userID, ok := userIDVal.(uint)
		if !ok {
			httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		if ownerID, found := getOwnerID(c); found && ownerID == userID {
			c.Next()
			return
		}

		roleValue, _ := c.Get("role")
		roleName, _ := roleValue.(string)
		if roleName != "" {
			allowed, err := checker.HasPermission(roleName, permission)
			if err == nil && allowed {
				c.Next()
				return
			}
		}

		httpx.Error(c, http.StatusForbidden, "Forbidden")
		c.Abort()
	}
}
