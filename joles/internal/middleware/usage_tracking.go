package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"lio-ai/internal/models"
	"lio-ai/internal/services"
)

// UsageTracking middleware tracks API usage automatically
func UsageTracking(usageService *services.UsageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip tracking for health and status endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/" {
			c.Next()
			return
		}

		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		durationMs := time.Since(startTime).Milliseconds()

		// Extract user ID from context or query params
		userID := c.GetString("user_id")
		if userID == "" {
			userID = c.Query("user_id")
		}
		if userID == "" {
			// Skip tracking if no user ID found
			return
		}

		// Determine request type based on endpoint
		requestType := determineRequestType(c.Request.URL.Path)
		if requestType == "" {
			// Skip tracking for non-tracked endpoints
			return
		}

		// Get token usage from context (should be set by handlers)
		tokensInput := c.GetInt("tokens_input")
		tokensOutput := c.GetInt("tokens_output")
		modelUsed := c.GetString("model_used")
		resourceID := c.GetInt64("resource_id")

		// Default model if not set
		if modelUsed == "" {
			modelUsed = "default"
		}

		// Track successful request
		success := c.Writer.Status() < 400
		errorMessage := ""
		if !success && c.Errors.Last() != nil {
			errorMessage = c.Errors.Last().Error()
		}

		// Create usage request
		usageReq := &models.UsageRequest{
			UserID:       userID,
			RequestType:  requestType,
			ResourceID:   resourceID,
			TokensInput:  tokensInput,
			TokensOutput: tokensOutput,
			ModelUsed:    modelUsed,
			DurationMs:   durationMs,
			Endpoint:     c.Request.URL.Path,
			Success:      success,
			ErrorMessage: errorMessage,
		}

		// Track usage asynchronously to avoid blocking response
		go func() {
			_ = usageService.TrackUsage(usageReq)
		}()
	}
}

// determineRequestType determines the request type based on endpoint path
func determineRequestType(path string) string {
	switch {
	case contains(path, "/chat"):
		return "chat"
	case contains(path, "/code"):
		return "code_generation"
	case contains(path, "/document"):
		return "document"
	default:
		return ""
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		len(s) > len(substr) && s[1:len(substr)+1] == substr
}

// QuotaCheck middleware checks if user has enough quota before processing
func QuotaCheck(usageService *services.UsageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip quota check for health and status endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/" {
			c.Next()
			return
		}

		// Extract user ID
		userID := c.GetString("user_id")
		if userID == "" {
			userID = c.Query("user_id")
		}
		if userID == "" {
			// Skip check if no user ID found
			c.Next()
			return
		}

		// Estimate tokens needed (conservative estimate)
		// This can be overridden by setting "tokens_needed" in context before this middleware
		tokensNeeded := c.GetInt("tokens_needed")
		if tokensNeeded == 0 {
			tokensNeeded = 4000 // Default estimate for typical request
		}

		modelUsed := c.GetString("model_used")
		if modelUsed == "" {
			modelUsed = "default"
		}

		// Check quota
		hasQuota, err := usageService.CheckQuota(userID, tokensNeeded, modelUsed)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to check quota: " + err.Error()})
			c.Abort()
			return
		}

		if !hasQuota {
			c.JSON(429, gin.H{
				"error": "quota exceeded",
				"message": "You have exceeded your daily or monthly token/cost limit. Please try again later or contact support to increase your quota.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
