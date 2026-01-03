package handlers

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// ProxyHandler proxies requests to the Python FastAPI service.
type ProxyHandler struct {
	targetURL string
	client    *http.Client
}

// NewProxyHandler creates a new proxy handler.
func NewProxyHandler(targetURL string) *ProxyHandler {
	return &ProxyHandler{
		targetURL: targetURL,
		client:    &http.Client{},
	}
}

// ProxyRequest proxies an HTTP request to the backend service.
func (ph *ProxyHandler) ProxyRequest(c *gin.Context) {
	// Block sensitive endpoints from being proxied
	blockedPaths := []string{"/docs", "/openapi.json", "/redoc"}
	for _, blocked := range blockedPaths {
		if c.Request.URL.Path == blocked {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Not found",
			})
			return
		}
	}

	// Build target URL - preserve query parameters
	targetURL := ph.targetURL + c.Request.URL.Path
	
	// Add user_id from JWT to query parameters if authenticated
	query := c.Request.URL.Query()
	if userID, exists := c.Get("user_id"); exists {
		// Add user_id to query string for backend API
		query.Set("user_id", userID.(string))
	}
	
	// Reconstruct URL with query parameters
	if len(query) > 0 {
		targetURL += "?" + query.Encode()
	}

	// Create new request
	proxyReq, err := http.NewRequest(
		c.Request.Method,
		targetURL,
		c.Request.Body,
	)
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create proxy request",
		})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Send request
	resp, err := ph.client.Do(proxyReq)
	if err != nil {
		log.Printf("Error proxying request: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Failed to reach backend service",
		})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Copy response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response",
		})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// HealthCheck checks both gateway and backend health.
func (ph *ProxyHandler) HealthCheck(c *gin.Context) {
	// Check backend health
	healthURL := ph.targetURL + "/health"
	resp, err := ph.client.Get(healthURL)
	backendStatus := "down"
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			backendStatus = "up"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"gateway": "up",
		"backend": backendStatus,
		"timestamp": os.Getenv("TIMESTAMP"),
	})
}
