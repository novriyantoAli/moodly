package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/moodly/internal/pkg/jwt"
	"github.com/novriyantoAli/moodly/internal/security"
	"github.com/novriyantoAli/moodly/internal/shared/apperror"
	"github.com/novriyantoAli/moodly/internal/shared/response"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
		)
	}
}

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal domain error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func JWTMiddleware(jwtManager *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			status, resp := apperror.ToHTTP(apperror.Unauthorized("token is required"))
			c.AbortWithStatusJSON(status, response.Response{
				Success: false,
				Error:   resp,
			})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtManager.ValidateToken(tokenStr)
		if err != nil {
			status, resp := apperror.ToHTTP(err)
			c.AbortWithStatusJSON(status, response.Response{
				Success: false,
				Error:   resp,
			})
			return
		}

		principal := security.Principal{
			UserID:      claims.UserID,
			Roles:       claims.Roles,
			Permissions: claims.Permissions,
		}

		ctx := security.WithPrincipal(c.Request.Context(), principal)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
