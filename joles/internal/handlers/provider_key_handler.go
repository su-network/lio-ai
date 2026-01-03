package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

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
	// Get authenticated user from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	keys, err := h.repo.GetAllByUser(userID.(string))
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
	// Get authenticated user from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	var req models.ProviderAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate required fields
	if req.Provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}
	if req.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "api_key is required"})
		return
	}

	// Convert models_enabled to JSON string
	modelsJSON := "[]"
	if len(req.ModelsEnabled) > 0 {
		b, _ := json.Marshal(req.ModelsEnabled)
		modelsJSON = string(b)
	}

	key := &models.ProviderAPIKey{
		UserID:        userID.(string),
		Provider:      req.Provider,
		APIKey:        req.APIKey,
		ModelsEnabled: modelsJSON,
	}

	if err := h.repo.Create(key); err != nil {
		// Log the actual error for debugging
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save API key",
			"details": err.Error(),
		})
		return
	}

	// Notify Python backend to reload models with new API keys
	go h.syncAPIKeysToBackend(userID.(string))

	c.JSON(http.StatusOK, gin.H{
		"message":  "API key saved successfully",
		"provider": req.Provider,
	})
}

// syncAPIKeysToBackend sends all user's API keys to Python backend
func (h *ProviderKeyHandler) syncAPIKeysToBackend(userID string) {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8000"
	}

	// Fetch all active API keys for this user
	keyResponses, err := h.repo.GetAllByUser(userID)
	if err != nil {
		log.Printf("Failed to fetch API keys for sync: %v", err)
		return
	}

	// Build API keys map - need to fetch decrypted keys
	apiKeys := make(map[string]string)
	for _, keyResp := range keyResponses {
		if keyResp.IsActive {
			// Fetch the actual decrypted key
			fullKey, err := h.repo.GetByUserAndProvider(userID, keyResp.Provider)
			if err != nil {
				log.Printf("Failed to fetch key for %s: %v", keyResp.Provider, err)
				continue
			}
			if fullKey != nil {
				apiKeys[fullKey.Provider] = fullKey.APIKey
			}
		}
	}

	// Send to Python backend
	payload := map[string]interface{}{
		"user_id":  userID,
		"api_keys": apiKeys,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(
		backendURL+"/api/v1/models/sync-keys",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		log.Printf("Failed to sync API keys to backend: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Printf("✓ API keys synced to Python backend for user %s", userID)
	} else {
		log.Printf("Failed to sync API keys: HTTP %d", resp.StatusCode)
	}
}

// DeleteKey soft deletes a provider API key
func (h *ProviderKeyHandler) DeleteKey(c *gin.Context) {
	// Get authenticated user from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	provider := c.Param("provider")

	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	if err := h.repo.Delete(userID.(string), provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}

	// Sync to Python backend to remove the provider
	go h.syncAPIKeysToBackend(userID.(string))

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
	// Get authenticated user from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	provider := c.Param("provider")

	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	key, err := h.repo.GetByUserAndProvider(userID.(string), provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch API key"})
		return
	}

	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	// Update last used
	h.repo.UpdateLastUsed(userID.(string), provider)

	c.JSON(http.StatusOK, gin.H{
		"provider": key.Provider,
		"api_key":  key.APIKey, // Only return decrypted key for internal use
	})
}

// SyncAllKeys manually syncs all user's API keys to Python backend
func (h *ProviderKeyHandler) SyncAllKeys(c *gin.Context) {
	// Get authenticated user from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Trigger sync in background
	go h.syncAPIKeysToBackend(userID.(string))

	c.JSON(http.StatusOK, gin.H{
		"message": "API keys sync triggered",
	})
}
