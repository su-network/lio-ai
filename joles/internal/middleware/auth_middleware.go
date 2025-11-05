package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/utils"
)

// AuthMiddleware handles authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Try Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// For now, we'll be permissive and allow requests without API keys
		// In production, you would validate the API key here
		if apiKey != "" {
			// TODO: Validate API key against database
			// For now, just extract user ID from key or use a default
			c.Set("authenticated", true)
			c.Set("api_key", apiKey)
		}

		// Extract user_id from query params as fallback
		userID := c.Query("user_id")
		if userID != "" {
			c.Set("user_id", userID)
		}

		c.Next()
	}
}

// RequireAuth middleware that enforces authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authenticated := c.GetBool("authenticated")
		if !authenticated {
			apiKey := c.GetHeader("X-API-Key")
			if apiKey == "" {
				authHeader := c.GetHeader("Authorization")
				if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if apiKey == "" {
				utils.UnauthorizedError(c, "API key required")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// AdminOnly middleware that requires admin role
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("user_role")
		if role != "admin" {
			utils.ForbiddenError(c, "Admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
