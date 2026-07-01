package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/moodly/internal/application/authorization/service"
	"github.com/novriyantoAli/moodly/internal/security"
	"github.com/novriyantoAli/moodly/internal/shared/apperror"
	"github.com/novriyantoAli/moodly/internal/shared/response"
	"go.uber.org/zap"
)

// RequirePermission checks if the authenticated user has the required permission
func RequirePermission(authService service.AuthorizationService, requiredPermission string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// // Get claims from context (set by AuthMiddleware)
		// claimsValue, exists := c.Get(string(jwt.ClaimsKey))
		// if !exists {
		// 	logger.Warn("Unauthorized access attempt: missing JWT claims")
		// 	c.JSON(http.StatusUnauthorized, gin.H{
		// 		"error":   "Unauthorized",
		// 		"message": "Authentication required",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// claims, ok := claimsValue.(*jwt.Claims)
		// if !ok {
		// 	logger.Error("Failed to cast claims from context")
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error":   "Internal Server Error",
		// 		"message": "Error processing authentication context",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		p, ok := security.PrincipalFromContext(c.Request.Context())
		if !ok {
			logger.Error("Failed to cast principal from context")

			status, resp := apperror.ToHTTP(apperror.BadRequest("Authentication required"))

			c.JSON(status, response.Response{
				Success: false,
				Error:   resp,
			})

			c.Abort()

			return
		}

		permissions, err := authService.GetPermissionsByRoles(c.Request.Context(), p.Roles)
		if err != nil {
			logger.Error("Failed to fetch permissions", zap.Error(err), zap.Strings("user_roles", p.Roles))

			status, resp := apperror.ToHTTP(apperror.Internal(err))
			c.JSON(status, response.Response{
				Success: false,
				Error:   resp,
			})

			c.Abort()

			return
		}

		// Check if required permission is in the list
		hasPermission := false
		for _, permission := range permissions {
			if permission == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			logger.Warn(
				fmt.Sprintf("Forbidden access attempt: user %d lacks permission %s", p.UserID, requiredPermission),
			)

			status, resp := apperror.ToHTTP(apperror.Forbidden("You don't have sufficient permissions to access this resource"))
			c.JSON(status, response.Response{
				Success: false,
				Error:   resp,
			})

			c.Abort()
			return
		}

		// set permissions to principal
		security.WithPrincipal(c.Request.Context(), security.Principal{
			UserID:      p.UserID,
			Roles:       p.Roles,
			Permissions: permissions,
		})

		// if !security.HasPermission(c.Request.Context(), requiredPermission) {
		// 	logger.Warn(
		// 		fmt.Sprintf("Forbidden access attempt: user %d lacks permission %s", p.UserID, requiredPermission),
		// 	)
		// 	c.JSON(http.StatusForbidden, gin.H{
		// 		"error":   "Forbidden",
		// 		"message": "You don't have sufficient permissions to access this resource",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// // Retrieve roles from claims
		// roles := claims.Roles
		// if len(roles) == 0 {
		// 	logger.Warn(fmt.Sprintf("Forbidden access attempt: user %d has no roles", claims.UserID))
		// 	c.JSON(http.StatusForbidden, gin.H{
		// 		"error":   "Forbidden",
		// 		"message": "You don't have sufficient permissions to access this resource",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// // Fetch permissions for the roles from repository (cached via Redis)
		// permissions, err := authService.GetPermissionsByRoles(c.Request.Context(), roles)
		// if err != nil {
		// 	logger.Error("Failed to fetch permissions", zap.Error(err), zap.Strings("user_roles", roles))
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error":   "Internal Server Error",
		// 		"message": "Error verifying permissions",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// // Check if required permission is in the list
		// hasPermission := false
		// for _, p := range permissions {
		// 	if p == requiredPermission {
		// 		hasPermission = true
		// 		break
		// 	}
		// }

		// if !hasPermission {
		// 	logger.Warn(
		// 		fmt.Sprintf("Forbidden access attempt: user %d lacks permission %s", claims.UserID, requiredPermission),
		// 	)
		// 	c.JSON(http.StatusForbidden, gin.H{
		// 		"error":   "Forbidden",
		// 		"message": "You don't have sufficient permissions to access this resource",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// User has permission, proceed to next handler
		c.Next()
	}
}

// RequireRoles checks if the authenticated user has at least one of the required roles
func RequireRoles(requiredRoles []string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context (set by AuthMiddleware)
		// claimsValue, exists := c.Get(string(jwt.ClaimsKey))
		// if !exists {
		// 	logger.Warn("Unauthorized access attempt: missing JWT claims")
		// 	c.JSON(http.StatusUnauthorized, gin.H{
		// 		"error":   "Unauthorized",
		// 		"message": "Authentication required",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// claims, ok := claimsValue.(*jwt.Claims)
		// if !ok {
		// 	logger.Error("Failed to cast claims from context")
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error":   "Internal Server Error",
		// 		"message": "Error processing authentication context",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// // Retrieve roles from claims
		// userRoles := claims.Roles
		// if len(userRoles) == 0 {
		// 	logger.Warn(fmt.Sprintf("Forbidden access attempt: user %d has no roles", claims.UserID))
		// 	c.JSON(http.StatusForbidden, gin.H{
		// 		"error":   "Forbidden",
		// 		"message": "You don't have sufficient permissions to access this resource",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// // Check if user has at least one of the required roles
		// hasRole := false
		// for _, requiredRole := range requiredRoles {
		// 	for _, userRole := range userRoles {
		// 		if userRole == requiredRole {
		// 			hasRole = true
		// 			break
		// 		}
		// 	}
		// 	if hasRole {
		// 		break
		// 	}
		// }

		// if !hasRole {
		// 	logger.Warn(
		// 		fmt.Sprintf("Forbidden access attempt: user %d lacks required roles %v", claims.UserID, requiredRoles),
		// 	)
		// 	c.JSON(http.StatusForbidden, gin.H{
		// 		"error":   "Forbidden",
		// 		"message": "You don't have sufficient roles to access this resource",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// User has role, proceed to next handler

		if !security.HasAnyIntersectionRoles(c.Request.Context(), requiredRoles) {
			logger.Warn(
				fmt.Sprintf("Forbidden access attempt:  lacks role %v", requiredRoles),
			)

			status, resp := apperror.ToHTTP(apperror.Forbidden("You don't have sufficient roles to access this resource"))
			c.JSON(status, response.Response{
				Success: false,
				Error:   resp,
			})

			c.Abort()
			return
		}

		// p, ok := security.PrincipalFromContext(c.Request.Context())
		// if !ok {
		// 	logger.Error("Failed to cast principal from context")
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error":   "Internal Server Error",
		// 		"message": "Error processing authentication context",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// for _, role := range requiredRoles {
		// 	if !security.HasRole(c.Request.Context(), role) {
		// 		logger.Warn(
		// 			fmt.Sprintf("Forbidden access attempt: user %d lacks role %s", p.UserID, role),
		// 		)
		// 		c.JSON(http.StatusForbidden, gin.H{
		// 			"error":   "Forbidden",
		// 			"message": "You don't have sufficient roles to access this resource",
		// 		})
		// 		c.Abort()
		// 		return
		// 	}
		// }
		c.Next()
	}
}
