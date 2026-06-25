package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/moodly/internal/application/security/repository"
	"github.com/novriyantoAli/moodly/internal/pkg/jwt"
	"go.uber.org/zap"
)

// RequirePermission checks if the authenticated user has the required permission
func RequirePermission(authRepo repository.AuthorizationRepository, requiredPermission string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context (set by AuthMiddleware)
		claimsValue, exists := c.Get(string(jwt.ClaimsKey))
		if !exists {
			logger.Warn("Unauthorized access attempt: missing JWT claims")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		claims, ok := claimsValue.(*jwt.Claims)
		if !ok {
			logger.Error("Failed to cast claims from context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Error processing authentication context",
			})
			c.Abort()
			return
		}

		// Retrieve roles from claims
		roles := claims.Roles
		if len(roles) == 0 {
			logger.Warn(fmt.Sprintf("Forbidden access attempt: user %d has no roles", claims.UserID))
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You don't have sufficient permissions to access this resource",
			})
			c.Abort()
			return
		}

		// Fetch permissions for the roles from repository (cached via Redis)
		permissions, err := authRepo.GetPermissionsByRoles(c.Request.Context(), roles)
		if err != nil {
			logger.Error("Failed to fetch permissions", zap.Error(err), zap.Strings("user_roles", roles))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Error verifying permissions",
			})
			c.Abort()
			return
		}

		// Check if required permission is in the list
		hasPermission := false
		for _, p := range permissions {
			if p == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			logger.Warn(
				fmt.Sprintf("Forbidden access attempt: user %d lacks permission %s", claims.UserID, requiredPermission),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You don't have sufficient permissions to access this resource",
			})
			c.Abort()
			return
		}

		// User has permission, proceed to next handler
		c.Next()
	}
}
