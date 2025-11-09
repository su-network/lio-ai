package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/models"
	"lio-ai/internal/utils"
)

// SystemHandler handles system-related requests
type SystemHandler struct {
	db        *sql.DB
	startTime time.Time
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(db *sql.DB) *SystemHandler {
	return &SystemHandler{
		db:        db,
		startTime: time.Now(),
	}
}

// HealthCheck performs a comprehensive health check
func (h *SystemHandler) HealthCheck(c *gin.Context) {
	checks := make(map[string]string)

	// Check database
	dbStatus := "up"
	if err := h.db.Ping(); err != nil {
		dbStatus = "down"
	}
	checks["database"] = dbStatus

	// Calculate uptime
	uptime := time.Since(h.startTime).String()

	response := models.HealthResponse{
		Status:    "operational",
		Gateway:   "up",
		Backend:   "up", // Will be updated by proxy handler
		Database:  dbStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "0.1.0",
		Uptime:    uptime,
		Checks:    checks,
	}

	// Overall status based on critical services
	if dbStatus == "down" {
		response.Status = "degraded"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetMetrics returns system metrics
func (h *SystemHandler) GetMetrics(c *gin.Context) {
	// Get total users
	var totalUsers, activeUsers int
	h.db.QueryRow("SELECT COUNT(*) FROM user_quotas").Scan(&totalUsers)
	h.db.QueryRow("SELECT COUNT(DISTINCT user_id) FROM usage_metrics WHERE created_at >= datetime('now', '-24 hours')").Scan(&activeUsers)

	// Get total chats and documents
	var totalChats, totalDocs int
	h.db.QueryRow("SELECT COUNT(*) FROM chats").Scan(&totalChats)
	h.db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&totalDocs)

	// Get usage metrics
	var totalRequests, successfulRequests, failedRequests int64
	var totalTokens int
	var totalCost float64
	var avgLatency float64

	h.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful,
			SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failed,
			COALESCE(SUM(tokens_total), 0) as tokens,
			COALESCE(SUM(cost_usd), 0.0) as cost,
			COALESCE(AVG(duration_ms), 0.0) as avg_latency
		FROM usage_metrics
	`).Scan(&totalRequests, &successfulRequests, &failedRequests, &totalTokens, &totalCost, &avgLatency)

	// Get endpoint statistics
	rows, err := h.db.Query(`
		SELECT 
			endpoint,
			COUNT(*) as request_count,
			AVG(duration_ms) as avg_time,
			CAST(SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) AS REAL) / COUNT(*) * 100 as error_rate
		FROM usage_metrics
		GROUP BY endpoint
		ORDER BY request_count DESC
		LIMIT 10
	`)
	
	var endpointStats []models.EndpointStat
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat models.EndpointStat
			rows.Scan(&stat.Endpoint, &stat.RequestCount, &stat.AverageTimeMs, &stat.ErrorRate)
			endpointStats = append(endpointStats, stat)
		}
	}

	// Get model statistics
	modelRows, err := h.db.Query(`
		SELECT 
			model_used,
			COUNT(*) as request_count,
			SUM(tokens_total) as total_tokens,
			SUM(cost_usd) as total_cost
		FROM usage_metrics
		WHERE model_used != ''
		GROUP BY model_used
		ORDER BY request_count DESC
		LIMIT 10
	`)

	var modelStats []models.ModelStat
	if err == nil {
		defer modelRows.Close()
		for modelRows.Next() {
			var stat models.ModelStat
			modelRows.Scan(&stat.ModelName, &stat.RequestCount, &stat.TotalTokens, &stat.TotalCostUSD)
			modelStats = append(modelStats, stat)
		}
	}

	metrics := models.MetricsResponse{
		RequestsTotal:      totalRequests,
		RequestsSuccessful: successfulRequests,
		RequestsFailed:     failedRequests,
		AverageLatencyMs:   avgLatency,
		ActiveUsers:        activeUsers,
		TotalUsers:         totalUsers,
		TotalChats:         totalChats,
		TotalDocuments:     totalDocs,
		TotalTokensUsed:    totalTokens,
		TotalCostUSD:       totalCost,
		EndpointStats:      endpointStats,
		ModelStats:         modelStats,
	}

	utils.SuccessResponse(c, metrics)
}

// GetInfo returns API information
func (h *SystemHandler) GetInfo(c *gin.Context) {
	info := gin.H{
		"name":        "Lio AI Gateway",
		"version":     "0.1.0",
		"description": "AI-powered code generation and chat API gateway",
		"uptime":      time.Since(h.startTime).String(),
		"features": []string{
			"Chat API",
			"Code Generation",
			"Document Management",
			"Usage Tracking",
			"Cost Monitoring",
			"RAG Search",
		},
		"endpoints": gin.H{
			"health":    "/health",
			"metrics":   "/api/v1/metrics",
			"documents": "/api/v1/documents",
			"chats":     "/api/v1/chats",
			"usage":     "/api/v1/usage",
			"codegen":   "/api/v1/codegen",
		},
	}

	utils.SuccessResponse(c, info)
}

// GetStats returns quick statistics
func (h *SystemHandler) GetStats(c *gin.Context) {
	var totalChats, totalDocs, totalMessages int
	h.db.QueryRow("SELECT COUNT(*) FROM chats").Scan(&totalChats)
	h.db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&totalDocs)
	h.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&totalMessages)

	var totalRequests int64
	var totalTokens int
	var totalCost float64
	h.db.QueryRow(`
		SELECT 
			COUNT(*),
			COALESCE(SUM(tokens_total), 0),
			COALESCE(SUM(cost_usd), 0.0)
		FROM usage_metrics
	`).Scan(&totalRequests, &totalTokens, &totalCost)

	stats := gin.H{
		"chats":         totalChats,
		"documents":     totalDocs,
		"messages":      totalMessages,
		"api_requests":  totalRequests,
		"tokens_used":   totalTokens,
		"total_cost_usd": totalCost,
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	utils.SuccessResponse(c, stats)
}
