package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lio-ai/internal/models"
	"lio-ai/internal/services"
)

// UsageHandler handles usage-related HTTP requests
type UsageHandler struct {
	usageService *services.UsageService
}

// NewUsageHandler creates a new usage handler
func NewUsageHandler(usageService *services.UsageService) *UsageHandler {
	return &UsageHandler{
		usageService: usageService,
	}
}

// GetQuotaStatus retrieves the current quota status
// GET /api/v1/usage/quota
func (h *UsageHandler) GetQuotaStatus(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	status, err := h.usageService.GetQuotaStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetUsageSummary retrieves aggregated usage statistics
// GET /api/v1/usage/summary
func (h *UsageHandler) GetUsageSummary(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	period := c.DefaultQuery("period", "monthly")
	if period != "daily" && period != "monthly" && period != "all_time" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period must be 'daily', 'monthly', or 'all_time'"})
		return
	}

	summary, err := h.usageService.GetUsageSummary(userID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// TrackUsage manually tracks a usage event (internal endpoint)
// POST /api/v1/usage/track
func (h *UsageHandler) TrackUsage(c *gin.Context) {
	var req models.UsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usageService.TrackUsage(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "usage tracked successfully"})
}

// CheckQuota checks if user has enough quota for a request
// POST /api/v1/usage/check-quota
func (h *UsageHandler) CheckQuota(c *gin.Context) {
	var req struct {
		UserID       string `json:"user_id" binding:"required"`
		TokensNeeded int    `json:"tokens_needed" binding:"required"`
		ModelName    string `json:"model_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hasQuota, err := h.usageService.CheckQuota(req.UserID, req.TokensNeeded, req.ModelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_quota": hasQuota,
		"user_id":   req.UserID,
		"tokens_needed": req.TokensNeeded,
	})
}

// UpdateQuota updates quota limits for a user
// PUT /api/v1/usage/quota/:user_id
func (h *UsageHandler) UpdateQuota(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var req models.QuotaUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usageService.UpdateQuota(userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "quota updated successfully"})
}

// GetDashboard returns a comprehensive dashboard of usage metrics
// GET /api/v1/usage/dashboard
func (h *UsageHandler) GetDashboard(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Get quota status
	quotaStatus, err := h.usageService.GetQuotaStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get daily summary
	dailySummary, err := h.usageService.GetUsageSummary(userID, "daily")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get monthly summary
	monthlySummary, err := h.usageService.GetUsageSummary(userID, "monthly")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dashboard := gin.H{
		"user_id":         userID,
		"quota_status":    quotaStatus,
		"daily_summary":   dailySummary,
		"monthly_summary": monthlySummary,
	}

	c.JSON(http.StatusOK, dashboard)
}
