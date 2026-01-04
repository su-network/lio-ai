package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/auth"
)

// NewAuthMiddleware creates authentication middleware with JWT validation
func NewAuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth validation for public auth endpoints (login, register)
		// These endpoints should not validate tokens at all
		if isPublicAuthEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get token from Authorization header or cookie
		token := ""

		// Check Authorization header first (Bearer token)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Fall back to cookie if no Authorization header
		if token == "" {
			var err error
			token, err = c.Cookie("auth_token")
			if err != nil {
				// No token found, continue without auth
				// (endpoint handler will decide if auth is required)
				c.Next()
				return
			}
		}

		// If still no token after checking both sources, continue without auth
		if token == "" {
			c.Next()
			return
		}

		// Validate JWT token (only if token exists)
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{
				"error": "invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Set claims in context for use in handlers
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("authenticated", true)

		c.Next()
	}
}

// RequireAuth middleware that enforces authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authenticated, exists := c.Get("authenticated")
		if !exists || !authenticated.(bool) {
			c.JSON(401, gin.H{
				"error": "authentication required",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware that requires specific role
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure authenticated
		authenticated, exists := c.Get("authenticated")
		if !exists || !authenticated.(bool) {
			c.JSON(401, gin.H{
				"error": "authentication required",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// Check roles
		rolesInterface, exists := c.Get("roles")
		if !exists {
			c.JSON(403, gin.H{
				"error": "insufficient permissions",
				"code":  "FORBIDDEN",
			})
			c.Abort()
			return
		}

		userRoles := rolesInterface.([]string)
		hasRole := false
		for _, userRole := range userRoles {
			for _, required := range requiredRoles {
				if userRole == required {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(403, gin.H{
				"error": "insufficient permissions",
				"code":  "FORBIDDEN",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isPublicAuthEndpoint checks if the path is a public authentication endpoint
// that should bypass auth validation completely
func isPublicAuthEndpoint(path string) bool {
	publicEndpoints := []string{
		"/api/v1/auth/register",
		"/api/v1/auth/login",
	}

	for _, endpoint := range publicEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}
	return false
}
