package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/models"
	"lio-ai/internal/repositories"
)

// ProviderKeyHandler handles provider API key operations
type ProviderKeyHandler struct {
	repo *repositories.ProviderKeyRepository
}

// NewProviderKeyHandler creates a new provider key handler
func NewProviderKeyHandler(repo *repositories.ProviderKeyRepository) *ProviderKeyHandler {
	return &ProviderKeyHandler{repo: repo}
}

// GetAllKeys gets all provider API keys for the current user
func (h *ProviderKeyHandler) GetAllKeys(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	keys, err := h.repo.GetAllByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch API keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys": keys,
	})
}

// CreateOrUpdateKey creates or updates a provider API key
func (h *ProviderKeyHandler) CreateOrUpdateKey(c *gin.Context) {
	var req models.ProviderAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Convert models_enabled to JSON string
	modelsJSON := "[]"
	if len(req.ModelsEnabled) > 0 {
		b, _ := json.Marshal(req.ModelsEnabled)
		modelsJSON = string(b)
	}

	key := &models.ProviderAPIKey{
		UserID:        userID,
		Provider:      req.Provider,
		APIKey:        req.APIKey,
		ModelsEnabled: modelsJSON,
	}

	if err := h.repo.Create(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "API key saved successfully",
		"provider": req.Provider,
	})
}

// DeleteKey soft deletes a provider API key
func (h *ProviderKeyHandler) DeleteKey(c *gin.Context) {
	provider := c.Param("provider")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if err := h.repo.Delete(userID, provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key deleted successfully",
	})
}

// HardDeleteKey permanently deletes a provider API key
func (h *ProviderKeyHandler) HardDeleteKey(c *gin.Context) {
	provider := c.Param("provider")
	userID := c.Query("user_id")
	
	// Optional: Add admin check here
	// For now, allow users to hard delete their own keys

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if err := h.repo.HardDelete(userID, provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to permanently delete API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key permanently deleted",
	})
}

// RestoreKey reactivates a soft-deleted provider API key
func (h *ProviderKeyHandler) RestoreKey(c *gin.Context) {
	provider := c.Param("provider")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if err := h.repo.Restore(userID, provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key restored successfully",
	})
}

// GetProviderKey retrieves the decrypted API key for a provider (internal use)
func (h *ProviderKeyHandler) GetProviderKey(c *gin.Context) {
	provider := c.Param("provider")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	key, err := h.repo.GetByUserAndProvider(userID, provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch API key"})
		return
	}

	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	// Update last used
	h.repo.UpdateLastUsed(userID, provider)

	c.JSON(http.StatusOK, gin.H{
		"provider": key.Provider,
		"api_key":  key.APIKey, // Only return decrypted key for internal use
	})
}
