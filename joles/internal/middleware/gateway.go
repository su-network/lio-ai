package middleware

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter implements token bucket rate limiting.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// AddClient adds a new client with specified rate limit.
func (rl *RateLimiter) AddClient(clientID string, rps float64, burst int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limiters[clientID] = rate.NewLimiter(rate.Limit(rps), burst)
}

// Allow checks if the request is allowed.
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.RLock()
	limiter, exists := rl.limiters[clientID]
	rl.mu.RUnlock()

	if !exists {
		// Default: 100 requests per second, burst of 10
		rl.AddClient(clientID, 100, 10)
		limiter, _ = rl.limiters[clientID]
	}

	return limiter.Allow()
}

// RateLimitMiddleware creates a Gin middleware for rate limiting.
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			c.JSON(429, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": 1,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoggingMiddleware logs incoming requests.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s %d (%s)",
			c.Request.Method,
			c.Request.RequestURI,
			c.ClientIP(),
			c.Writer.Status(),
			duration,
		)
	}
}

// CORSMiddleware enables CORS.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ErrorRecoveryMiddleware recovers from panics.
func ErrorRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(500, gin.H{
					"error": fmt.Sprintf("Internal server error: %v", err),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
