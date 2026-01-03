package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CSRFHeaderName = "X-CSRF-Token"
	CSRFCookieName = "_csrf"
)

// GenerateCSRFToken creates a new CSRF token
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// CSRFMiddleware protects against CSRF attacks
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF check for public auth endpoints (login, register)
		if isPublicAuthEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get or generate CSRF token
		token, err := c.Cookie(CSRFCookieName)
		if err != nil || token == "" {
			newToken, err := GenerateCSRFToken()
			if err != nil {
				c.JSON(500, gin.H{
					"error": "internal server error",
					"code":  "INTERNAL_ERROR",
				})
				c.Abort()
				return
			}

			log.Printf("[CSRF] Generating new CSRF token for path: %s", c.Request.URL.Path)

			// Set token in cookie (NOT httpOnly so JS can read it)
			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(
				CSRFCookieName,
				newToken,
				3600,
				"/",
				"",    // Empty domain works for same-origin (via proxy)
				false, // httpOnly - must be false so JavaScript can read it
				false, // secure - false for HTTP localhost
			)
			token = newToken
		} else {
			log.Printf("[CSRF] Using existing CSRF token for path: %s", c.Request.URL.Path)
		}

		// Store token in context for template use
		c.Set("csrf_token", token)

		// For state-changing requests, validate token
		if isStatefulRequest(c.Request.Method) {
			headerToken := c.GetHeader(CSRFHeaderName)
			if headerToken == "" {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "csrf token required",
					"code":  "CSRF_TOKEN_MISSING",
				})
				c.Abort()
				return
			}

			// URL-decode the header token if needed (replace %3D with =)
			headerToken = strings.ReplaceAll(headerToken, "%3D", "=")
			headerToken = strings.ReplaceAll(headerToken, "%2B", "+")
			headerToken = strings.ReplaceAll(headerToken, "%2F", "/")

			// Debug logging
			log.Printf("[CSRF] Cookie token: %s", token)
			log.Printf("[CSRF] Header token: %s", headerToken)
			log.Printf("[CSRF] Tokens match: %v", strings.EqualFold(token, headerToken))

			if !strings.EqualFold(token, headerToken) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "invalid csrf token",
					"code":  "CSRF_TOKEN_INVALID",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// isStatefulRequest checks if request modifies state
func isStatefulRequest(method string) bool {
	return method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH"
}

// isPublicAuthEndpoint checks if the path is a public authentication endpoint
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
